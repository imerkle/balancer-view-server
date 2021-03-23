package config

// Exchange ...
type Exchange struct {
	Value string `json:"value"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
}

// ChartConfig ...
type ChartConfig struct {
	Resolutions    []string   `json:"supported_resolutions"`
	Exchanges      []Exchange `json:"exchanges"`
	GroupRequest   bool       `json:"supports_group_request"`
	Marks          bool       `json:"supports_marks"`
	Search         bool       `json:"supports_search"`
	TimescaleMarks bool       `json:"supports_timescale_marks"`
	Time           bool       `json:"supports_time"`
}

// YamlConfig describes cofigurations input from yaml
type YamlConfig struct {
	ChartConfig struct {
		SupportedResolutions []string `yaml:"supported_resolutions"`
	} `yaml:"chart_config"`
	SyncInterval int64 `yaml:"sync_interval"`
	ResetDb      bool  `yaml:"reset_db"`
}
