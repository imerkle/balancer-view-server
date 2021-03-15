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
)

var Symbols []config.Symbol // Supported symbols

func InitSymbols() {
	log.Println("Updaing symbols")
	for k, _ := range PairsMap {
		s := config.NewSymbol(
			k,
			k,
			k)
		Symbols = append(Symbols, s)
	}
}

type Syncer struct {
	SwapTimestamp     int64
	TargetedTimestamp int64
	NumInserts        int
	FirstSwaps        int
	Batch             pgx.Batch
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
	PairsMap      = map[string]bool{}
	MajorQuotes   = map[string]bool{"USDC": true, "WETH": true}
	pairsMapMutex = sync.RWMutex{}
)

func (x *SyncerGroup) Init() {
	x.BatchSeconds = 60 * 60 * 24 * 7 * 4 //1 week
	//x.TargetedTimestamp = 1591979848
	x.TargetedTimestamp = time.Now().UTC().Unix()

	//setup clients
	db.InitGQL()

	//get last timestamp to reume syncing from
	var start int64 = 0
	rows, err := db.Dbpool.Query(context.Background(), `select time FROM swaps order by time desc limit 1`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching timestamp from db %v\n", err)
	}
	type result struct {
		time time.Time
	}
	rows.Next()
	var r result
	err = rows.Scan(&r.time)
	if err == nil {
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
	for {
		{
			type SQF struct {
				Swap []Swap `graphql:"swaps(first:1, orderBy: timestamp, orderDirection: desc)"`
			}
			var query SQF
			err = db.Gqlclient.Query(context.Background(), &query, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching first swap: %v\n", err)
				os.Exit(1)
			}
			start = int64(query.Swap[0].Timestamp)
		}
		//start parallel syncers
		var wg sync.WaitGroup
		end := start + x.BatchSeconds
		totalBatches := 0
		for start < x.TargetedTimestamp {
			totalBatches++
			wg.Add(1)
			go x.startSync(&wg, start, end)
			start = end
			end = start + x.BatchSeconds
		}
		fmt.Println("Total Batches " + strconv.Itoa(totalBatches))
		wg.Wait()
		fmt.Println("All Sync completed")
		time.Sleep(time.Duration(x.SyncInterval) * time.Second)
		x.TargetedTimestamp = time.Now().UTC().Unix()
	}
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
			PairsMap[r.Pair] = true
		}
	}
	InitSymbols()
	defer rows2.Close()
}
func (x *SyncerGroup) startSync(wg *sync.WaitGroup, start int64, end int64) {
	syncer := &Syncer{}
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
	x.Batch = pgx.Batch{}
	x.TargetedTimestamp = targetedTimestamp
	db.Gqlclient = graphql.NewClient("https://api.thegraph.com/subgraphs/name/balancer-labs/balancer", nil)
}

func (x *Syncer) Start() {
	_, currBatchNum := x.Batching()
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

//Batching fetch/batch
func (x *Syncer) Batching() (SwapQuery, int) {
	x.FetchSwaps()
	n := x.CreateBatch()
	return x.query, n
}

type SwapQuery struct {
	Swaps []Swap `graphql:"swaps(first:$first_swaps, orderBy: timestamp, where:{timestamp_gte: $swap_timestamp})"`
}

//FetchSwaps fetches swaps
func (x *Syncer) FetchSwaps() {
	x.query = SwapQuery{}
	variables := map[string]interface{}{
		"first_swaps":    graphql.Int(x.FirstSwaps),
		"swap_timestamp": graphql.Int(x.SwapTimestamp),
	}
	err := db.Gqlclient.Query(context.Background(), &x.query, variables)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Retrying...")
		x.FetchSwaps()
	}
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
			if _, ok := PairsMap[pair]; !ok {
				if _, ok := PairsMap[pairRev]; ok {
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
				pairsMapMutex.Lock()
				PairsMap[pair] = true
				pairsMapMutex.Unlock()
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
	br := db.Dbpool.SendBatch(context.Background(), &x.Batch)

	//execute statements in batch queue
	for i := 0; i < x.NumInserts; i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to execute statement in batch queue %v\n", err)
			return err
		}
	}
	fmt.Println("Insterted " + strconv.Itoa(x.NumInserts) + " swaps till " + time.Unix(x.SwapTimestamp, 0).String() + " ")
	return nil
}
