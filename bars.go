package main

import (
	"time"
)

// Bar as described here: https://github.com/tradingview/charting_library/wiki/UDF#bars
type Bar struct {
	Time         time.Time
	TimeEnd      time.Time
	UnixTime     int64   `json:"t"` // Unix Epoch time in seconds
	ClosingPrice float64 `json:"c"`
	OpeningPrice float64 `json:"o"`
	HighPrice    float64 `json:"h"`
	LowPrice     float64 `json:"l"`
	Volume       float64 `json:"v"`
}
