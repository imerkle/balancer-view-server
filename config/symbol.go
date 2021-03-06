package config

// Symbol describes a tradable entity for which trading chart can be generated.
// Attributes must be as exactly listed here:
// https://github.com/tradingview/charting_library/wiki/Symbology
type Symbol struct {
	// Unique identifier of the Symbol
	// Example: BTC for Bitcoin
	Ticker string `json:"ticker"`
	// Official name of the symbol
	// Example: Ethereum
	Name string `json:"name"`
	// Description of a symbol.
	// Will be displayed in the chart legend for this symbol.
	Description string `json:"description"`
	// Type of the Symbol.
	// Supported types:
	// stock, forex, futures, bitcoin etc or a custom string value
	// [=] use "bitcoin" or "coin/token"
	Type string `json:"type"`
	// Start and end time of trading daily sessions.
	// Format: HHMM-HHMM
	// Example: "09-05"
	// Exception: "24x7" for symbols that are traded 24/7
	// [=] use "24x7"
	Session string `json:"session"`
	// Both Exchange and ListedExchange fields are expected to have a
	// short name of the exchange where this symbol is traded.
	// The name will be displayed in the chart legend for this symbol.
	// [=] use "HaloDEX"
	Exchange string `json:"exchange-traded"`
	// [=] use "HaloDEX"
	ListedExchange string `json:"exchange-listed"`
	// Timezone of the exchange for this symbol in "olsondb" format.
	// [=] use "Etc/UTC"
	TimeZone string `json:"timezone"`
	// MinMov is the amount of price precision steps for 1 tick.
	// For example, since the tick size for U.S. equities is 0.01, minmov is 1.
	// But the price of the E-mini S&P futures contract moves upward
	// or downward by 0.25 increments, so the minmov is 25.
	// [~] start by using 0.01
	MinMov float64 `json:"minmov"`
	// PriceScale defines the number of decimal places.
	// It is 10^number-of-decimal-places.
	// If a price is displayed as 1.01, pricescale is 100;
	// If it is displayed as 1.005, pricescale is 1000.
	// [=] use 1e8
	PriceScale int64 `json:"pricescale"`
	// MinMove2 for common prices is 0 or it can be skipped
	// [=] use 0
	MinMove2 float64 `json:"minmov2"`
	// Fractional for common prices is false or it can be skipped
	// [=] use false
	Fractional bool `json:"fractional"`
	// Boolean value showing whether the symbol includes intraday (minutes)
	// historical data. If it's false then all buttons for intraday resolutions
	// will be disabled for this particular symbol. If it is set to true, all
	// resolutions that are supplied directly by the datafeed must be
	// provided in intraday_multipliers array.
	// Default: false
	// [?] not sure. trial and error
	HasIntraDay bool `json:"has_intraday"`
	// An array of resolutions which should be enabled for this symbol.
	// Resolution or time interval is a time period of one bar.
	// More information:
	// https://github.com/tradingview/charting_library/wiki/Resolution
	// https://github.com/tradingview/charting_library/wiki/JS-Api#supported_resolutions
	// Example: ["1", "15", "30", "60", "D", "6M", "Y"] will give you "1 minute, 15 minutes,
	// 30 minutes, 1 hour, 1 day, 6 months, 1 year" in resolution widget.
	// [~] start with ["5", "15", "30", "60"]
	SupportedResolutions []string `json:"supported_resolutions"`
	// Array of resolutions (in minutes) supported directly by the data feed.
	// The default of [] means that the data feed supports aggregating by any number of minutes.
	// https://github.com/tradingview/charting_library/wiki/Symbology#intraday_multipliers
	// [~] leave empty for now
	IntraDayMultipliers []string `json:"intraday_multipliers"`
	// Boolean value showing whether the symbol includes seconds in the historical data.
	// [~] ignore
	//HasSeconds bool `json:"has_seconds"`
	// An array containing resolutions that include seconds (excluding postfix)
	// that the data feed provides. E.g., if the data feed supports resolutions
	// such as ["1S", "5S", "15S"], but has 1-second bars for some symbols then
	// you should set seconds_multipliers of this symbol to [1].
	// This will make Charting Library build 5S and 15S resolutions by itself.
	// [-] ignore
	//SecondsMultipliers []string `json:"seconds_multipliers"`
	// Whether data feed has its own daily resolution bars or not.
	// [=] use false. TODO: generate daily bar data.
	//HasDaily bool `json:"has_daily"`
	// Whether data feed has its own weekly and monthly resolution bars or not.
	// [=] use false.
	//HasWeeklyAndMonthly bool `json:"has_weekly_and_monthly"`
	// whether the library should generate empty bars in the session
	// When there is no data from the data feed for this particular time.
	// [=] use true
	//HasEmptyBars bool `json:"has_empty_bars"`
	// Whether the library should filter bars using the current trading session.
	// If false, bars will be filtered only when the library builds
	// data from another resolution or if has_empty_bars was set to true.
	// If true, the Library will remove bars that don't belong to the trading session from data feed.
	// Default true
	// [=] use default true
	//ForceSessionRebuild bool `json:"force_session_rebuild"`
	// Boolean showing whether the symbol includes volume data or not.
	// [?] not sure
	//HasNoVolume bool `json:"has_no_volume"`
	// Integer showing typical volume value decimal places for a particular symbol.
	// 0 means volume is always an integer.
	//  1 means that there might be 1 numeric character after the comma.
	// [=] use 0
	//VolumePrecision int `json:"volume_precision"`
	// The status code of a series with this symbol. The status is shown in the upper right corner of a chart.
	// Supported statuses: streaming, endofday, pulsed, delayed_streaming
	// [=] use "pulsed" only
	//DataStatus string `json:"data_status"`
	// Whether this symbol is an expired futures contract or not.
	// [-] ignore
	//Expired bool `json:"expired"`
	// If Expired is set to true. UNIX epoch time in seconds
	// [-] ignore
	//ExpirationDate int64 `json:"expiration_date"`
	// [-] ignore
	//Sector string `json:"sector"`
	// [-] ignore
	//Industry string `json:"industry"`
	// [-] ignore ???
	//CurrencyCode string `json:"currency_code"`
}

// instantiate a Symbol struct with default values
func NewSymbol(name, ticker, description string) (s Symbol) {
	s.Name = name
	s.Ticker = ticker
	s.Description = description
	s.Type = "Crypto" // use token/coin
	s.Session = "24x7"
	s.Exchange = "Balancer"
	s.ListedExchange = "Balancer"
	s.TimeZone = "Etc/UTC"
	s.MinMov = 1
	s.PriceScale = 1e8
	s.SupportedResolutions = Conf.Resolutions
	s.HasIntraDay = true // [?]
	s.IntraDayMultipliers = []string{"1"}
	/*
		for i := 0; i < len(resolutionMins); i++ {
			s.IntraDayMultipliers = append(s.IntraDayMultipliers, fmt.Sprint(resolutionMins[i]))
		}

		s.HasDaily = false // [?]
		s.HasEmptyBars = false
		s.ForceSessionRebuild = true
		s.DataStatus = "pulsed"
		s.HasNoVolume = false
	*/
	return
}
