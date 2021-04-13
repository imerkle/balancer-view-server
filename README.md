# BalancerView Server

High performance go server for syncing balancer swaps and serving to tradingview client

### Details

This server has two parts. 

##### Syncer 

 It parallely sync all swaps from subgraph and stores in a timeseries db https://github.com/timescale/timescaledb

##### Universal DataFeed

A datafeed server for tradingview charting library that serves the data from timescale db

### usage

- `go run *.go`
- or use docker
- or deploy into kubernetes

### config 

default `config.yaml`

```
chart_config:
  supported_resolutions:  # add any custom resolution, e.g 240(4h), 420(7h)
    - "1"
    - "5"
    - "15"
    - "30"
    - "60"
    - "1D"
    - "1W"
    - "1M"
sync_interval: 60 #sync new swaps every second
batch_days: 100 #days for each batch during initial syncing . less days = more parallel workers
reset_db: false #true = reset db during start, enable for dev only
```

### improvements

- Create materialized views of candlesticks for high liquidity pairs (not required as of now as there aren't many swaps which server can't handle to serve on spot)

- generated automated `usdc` pairs for pairs that dont have `usdc` quote for easier pricing