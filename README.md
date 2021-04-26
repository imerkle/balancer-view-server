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
endpoints:
  - "https://api.thegraph.com/subgraphs/name/balancer-labs/balancer"
  - "https://api.thegraph.com/subgraphs/name/balancer-labs/balancer-v2"
```


### Changelog

- Ordered list quote instead of maps 
- Supports syncing multiple endpoints

### Roadmap

*just some ideas nothing set in stone*

- Better logs [I1]
- Materialized views of candlesticks for high liquidity pairs [I1]
- Integerate chart with exchange interface [F1]
- Automatically generate custom Fiat pairs for every tokens [F2]
- Custom price alerts to discord, slack etc [F3]
- Screener [F3]

*F- Feature*
*I- Internal*
*expect F1 to come sooner than F3*