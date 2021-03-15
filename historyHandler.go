package main

import (
	"balancer-view/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const historyStatusNoData = "no_data"
const historyStatusOk = "ok"
const historyStatusError = "error"

var resolutionCache map[string][]Bar
var cachedBars map[string]map[string][]Bar // Symbol : resolution : []Bar

// History as described here: https://github.com/tradingview/charting_library/wiki/UDF#bars
type History struct {
	// Valid statuses : ok | error | no_data
	Status string `json:"s"`
	// only when Status == error
	ErrorMessage string `json:"errmsg"`
	// Unix Epoch time in seconds
	BarTime      []int64   `json:"t"`
	ClosingPrice []float64 `json:"c"`
	OpeningPrice []float64 `json:"o"`
	HighPrice    []float64 `json:"h"`
	LowPrice     []float64 `json:"l"`
	Volume       []float64 `json:"v"`
	// Unix Epoch time of the next bar.
	// Only if status == no_data
	NextTime int64 `json:"nextTime"`
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	symbol := strings.ToLower(params["symbol"][0])
	resolution := params["resolution"][0]
	log.Println("Resolution:", resolution)
	from, _ := strconv.ParseInt(params["from"][0], 0, 64)
	to, _ := strconv.ParseInt(params["to"][0], 0, 64)
	h := History{}
	h.Status = historyStatusOk
	nextTime := int64(0)
	bars, err := getBars(symbol, resolution)
	if respondIfError(err, w, "Failed to read file or symbol not found", err500) {
		return
	}
	for i := 0; i < len(bars); i++ {
		t := bars[i].UnixTime
		if t >= from && t <= to {
			h.BarTime = append(h.BarTime, t)
			h.ClosingPrice = append(h.ClosingPrice, bars[i].ClosingPrice)
			h.OpeningPrice = append(h.OpeningPrice, bars[i].OpeningPrice)
			h.HighPrice = append(h.HighPrice, bars[i].HighPrice)
			h.LowPrice = append(h.LowPrice, bars[i].LowPrice)
			h.Volume = append(h.Volume, bars[i].Volume)
		}
	}

	if len(h.BarTime) == 0 {
		h.Status = historyStatusNoData
		h.NextTime = nextTime
	}
	respondJSON(w, h, ok200)
}

func normalizeResolution(resolution string) string {
	_, err := strconv.ParseInt(resolution, 10, 64)
	if err != nil {
		return resolution
	}
	return resolution + "m"
}
func getBars(symbol, resolution string) (bars []Bar, err error) {
	ctx := context.Background()
	connStr := utils.GetEnv("POSTGRES_CONN", utils.PostgresConn)
	dbpool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	query := `SELECT time_bucket($1, time) as t,first(price, time) AS open,last(price, time) AS close,MAX(price) AS high,MIN(price) AS low,SUM(amount) AS volume FROM swaps WHERE pair=$2 GROUP BY t, pair;`
	rows, err := dbpool.Query(ctx, query, normalizeResolution(resolution), strings.ToUpper(symbol))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching swaps from db %v\n", err)
	}
	type result struct {
		t      time.Time
		open   float64
		close  float64
		high   float64
		low    float64
		volume float64
	}
	for rows.Next() {
		var r result
		err = rows.Scan(&r.t, &r.open, &r.close, &r.high, &r.low, &r.volume)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to scan rows for swaps %v\n", err)
		} else {
			bars = append(bars, Bar{
				UnixTime:     r.t.Unix(),
				ClosingPrice: r.close,
				OpeningPrice: r.open,
				HighPrice:    r.high,
				LowPrice:     r.low,
				Volume:       r.volume,
			})
		}
	}
	defer rows.Close()
	return bars, nil
}
