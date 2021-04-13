package syncer

import (
	"balancer-view/config"
	"balancer-view/db"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/jackc/pgx/v4"
	cmap "github.com/orcaman/concurrent-map"
)

var Symbols []config.Symbol // Supported symbols

func InitSymbols() {
	log.Println("Updating symbols")
	for _, k := range PairsMap.Keys() {
		s := config.NewSymbol(
			k,
			k,
			k)
		Symbols = append(Symbols, s)
	}
}

type Syncer struct {
	ID                int
	SwapTimestamp     int64
	TargetedTimestamp int64
	NumInserts        int
	FirstSwaps        int
	Batch             *pgx.Batch
	query             SwapQuery
}

//Syncer syncs swaps from graphql indexer to timeseries
type SyncerGroup struct {
	//maunaly syncs till targetted timestamp then starts a subscription for realtime syncing
	TargetedTimestamp int64
	BatchSeconds      int64
	SyncInterval      int64
}

var (
	PairsMap    = cmap.New()
	MajorQuotes = map[string]bool{"USDC": true, "WETH": true}
)

func (x *SyncerGroup) Init(batchDays int64) {
	x.BatchSeconds = 60 * 60 * 24 * batchDays //1 week
	//x.TargetedTimestamp = 1591979848
	x.TargetedTimestamp = time.Now().UTC().Unix()

	//setup clients
	db.InitGQL()

	//get last timestamp to reume syncing from
	var start int64 = 0
	rows, err := db.Dbpool.Query(context.Background(), `select time FROM swaps order by time desc limit 1`)

	type result struct {
		time time.Time
	}
	rows.Next()
	var r result
	err = rows.Scan(&r.time)
	tt := time.Time{}
	if err == nil && r.time != tt {
		start = r.time.UTC().Unix()
	}
	defer rows.Close()

	//first time so get starting timestamp to start syncing
	if start == 0 {
		type SQF struct {
			Swap []Swap `graphql:"swaps(first:1, orderBy: timestamp)"`
		}
		var query SQF
		err = db.Gqlclient.Query(context.Background(), &query, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching first swap: %v\n", err)
			os.Exit(1)
		}
		start = int64(query.Swap[0].Timestamp)
	}

	//setup pairs
	SetupPairs()

	go func() {
		for range time.Tick(time.Duration(x.SyncInterval) * time.Second) {
			//start parallel syncers
			var wg sync.WaitGroup
			end := start + x.BatchSeconds
			totalBatches := 0
			fmt.Println("Sync started from")
			fmt.Println(start)
			for start < x.TargetedTimestamp {
				totalBatches++
				wg.Add(1)
				go x.startSync(totalBatches, &wg, start, end)
				start = end
				end = start + x.BatchSeconds
			}
			fmt.Println("Total Batches " + strconv.Itoa(totalBatches))
			wg.Wait()
			fmt.Println("All Sync completed")
			start = x.TargetedTimestamp
			x.TargetedTimestamp = time.Now().UTC().Unix()
		}
	}()
}
func SetupPairs() {
	rows2, err := db.Dbpool.Query(context.Background(), `select pair from pairs`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching pairs from db %v\n", err)
	}
	type result2 struct {
		Pair string
	}
	for rows2.Next() {
		var r result2
		err = rows2.Scan(&r.Pair)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to scan rows2 for Pairs %v\n", err)
		} else {
			PairsMap.Set(r.Pair, true)
		}
	}
	InitSymbols()
	defer rows2.Close()
}
func (x *SyncerGroup) startSync(id int, wg *sync.WaitGroup, start int64, end int64) {
	syncer := &Syncer{ID: id}
	syncer.Init(start, end)
	syncer.Start()
	wg.Done()
}

/*
func (x *Syncer) CreateMaterializedView(pair string) {

	pairview := pair + "_view"
	query1 := fmt.Sprintf("CREATE MATERIALIZED VIEW %s WITH (timescaledb.continuous) AS SELECT time_bucket('1h', time) as interval,first(price, time) AS open,last(price, time) AS close,MAX(price) AS high,MIN(price) AS low,SUM(amount) AS volume FROM swaps WHERE pair='%s' GROUP BY interval, pair;", pairview, pair)
	query2 := fmt.Sprintf("SELECT add_continuous_aggregate_policy('%s', start_offset => INTERVAL '1 week', end_offset   => INTERVAL '1 hour', schedule_interval => INTERVAL '30 minutes');", pairview)
	_, err1 := x.Dbpool.Exec(x.ctx, query1)
	_, err2 := x.Dbpool.Exec(x.ctx, query2)
	if err1 != nil {
		fmt.Fprintf(os.Stderr, "Unable to create Materialized view: %v\n", err1)
	}
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "Unable to create continious policy: %v\n", err1)
	}

}*/

//Init initializes Syncer
func (x *Syncer) Init(swapTimestamp int64, targetedTimestamp int64) {
	x.SwapTimestamp = swapTimestamp
	x.NumInserts = 0
	x.FirstSwaps = 1000
	x.Batch = &pgx.Batch{}
	x.TargetedTimestamp = targetedTimestamp
	db.Gqlclient = graphql.NewClient("https://api.thegraph.com/subgraphs/name/balancer-labs/balancer", nil)
}

func (x *Syncer) Start() {
	err := x.FetchSwaps()
	//fmt.Println("Starting for ID: ", x.ID, x.NumInserts)
	if err != nil {
		time.Sleep(5 * time.Second)
		fmt.Println(err)
		fmt.Println("Fetch failed sync retry for ID: ", x.ID, x.NumInserts)
		x.Start()
	} else {
		currBatchNum := x.CreateBatch()
		//fmt.Println("Fetched " + strconv.Itoa(x.Batch.Len()) + " " + strconv.Itoa(x.NumInserts) + " Swaps till " + time.Unix(x.SwapTimestamp, 0).String() + " ")
		if currBatchNum < x.FirstSwaps {
			err := x.InsertBatch()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Sync Completed till " + time.Unix(x.TargetedTimestamp, 0).String())
		} else {
			x.Start()
		}
	}
}

type SwapQuery struct {
	Swaps []Swap `graphql:"swaps(first:$first_swaps, orderBy: timestamp, where:{timestamp_gte: $swap_timestamp})"`
}

//FetchSwaps fetches swaps
func (x *Syncer) FetchSwaps() error {
	x.query = SwapQuery{}
	variables := map[string]interface{}{
		"first_swaps":    graphql.Int(x.FirstSwaps),
		"swap_timestamp": graphql.Int(x.SwapTimestamp),
	}
	err := db.Gqlclient.Query(context.Background(), &x.query, variables)
	return err
}

var queryInsertTimeseriesData = `INSERT INTO swaps (id, pair, price, amount, time) VALUES ($1, $2, $3, $4, $5);`
var queryInsertPair = `INSERT INTO pairs (pair) VALUES ($1);`

//CreateBatch creates batch from fetched swaps for insertion
func (x *Syncer) CreateBatch() int {
	var currBatchNum = 0
	for _, swap := range x.query.Swaps {
		if swap.Timestamp < graphql.Int(x.TargetedTimestamp) {
			/*
				LINKWETH
				first = LINK = TokenAmountIn
				second = WETH = TokenAmountOut
			*/
			pair := string(swap.TokenInSym) + string(swap.TokenOutSym)
			pairRev := string(swap.TokenOutSym) + string(swap.TokenInSym)
			insertPair := false
			isRev := false

			if _, ok := PairsMap.Get(pair); !ok {
				if _, ok := PairsMap.Get(pairRev); ok {
					pair = pairRev
					isRev = true
				} else {
					/*
						pair = WETHLINK
						pairrev = LINKWETH
					*/
					if _, ok := MajorQuotes[string(swap.TokenInSym)]; ok {
						pair = pairRev
						isRev = true
					}
					insertPair = true
				}
			}
			if insertPair {
				x.Batch.Queue(queryInsertPair, pair)
				PairsMap.Set(pair, true)
				Symbols = append(Symbols, config.NewSymbol(
					pair,
					pair,
					pair))
			}
			firstAmount, _ := strconv.ParseFloat(string(swap.TokenAmountIn), 64)
			secondAmount, _ := strconv.ParseFloat(string(swap.TokenAmountOut), 64)
			if isRev {
				//swap
				tmpAmount := secondAmount
				secondAmount = firstAmount
				firstAmount = tmpAmount
			}
			timestamp := time.Unix(int64(swap.Timestamp), 0)
			x.Batch.Queue(queryInsertTimeseriesData, swap.ID, pair, secondAmount/firstAmount, firstAmount, timestamp)
			x.NumInserts++
			currBatchNum++
			x.SwapTimestamp = int64(swap.Timestamp)
			if swap.ID == "" || secondAmount == 0.0 || firstAmount == 0.0 || swap.Timestamp == 0 {
				fmt.Println("Invalid Data")
				fmt.Println(swap)
			}
		}
	}
	return currBatchNum
}

//InsertBatch inserts into db
func (x *Syncer) InsertBatch() error {

	//send batch to connection pool
	fmt.Println("Inserting DB BATCH.... ID: ", x.ID)
	br := db.Dbpool.SendBatch(context.Background(), x.Batch)
	fmt.Println("DONE INSERTING DB BATCH.... ID: ", x.ID)

	//execute statements in batch queue
	for i := 0; i < x.NumInserts; i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to execute statement in batch queue %v\n", err)
			return err
		}
	}
	defer br.Close()
	fmt.Println("Insterted " + strconv.Itoa(x.NumInserts) + " swaps till " + time.Unix(x.SwapTimestamp, 0).String() + " ")
	return nil
}
