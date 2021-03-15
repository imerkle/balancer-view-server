package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var Conf ChartConfig
var Resolutions []string // sypported bar resolutions

func ConfInit(yconf YamlConfig) {
	Conf = ChartConfig{
		TimescaleMarks: false,
		Marks:          false,
		GroupRequest:   false,
		Search:         true,
		Time:           true,
		Resolutions:    yconf.ChartConfig.SupportedResolutions,
		Exchanges:      []Exchange{{Value: "", Name: "All Exchanges", Desc: ""}, {Value: "Balancer", Name: "Balancer", Desc: "Balancer"}},
	}
}
func (c *YamlConfig) GetConf(workspaceRoot string) {
	yamlFile, err := ioutil.ReadFile(workspaceRoot + "config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
