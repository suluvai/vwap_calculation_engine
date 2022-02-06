/*
Responsible for retrieving config relating to the vwap_calculation_engine application itself.
*/
package conf

import (
	"encoding/json"
	"log"
	"os"
)

var (
	Configuration Config
)

type Config struct {
	WebServerURL   string
	VwapWindowSize uint
	TradingPairs   []string
	LogLevel       string
}

func getConfigPath() string {
	const devConfigPath = "./conf/app.config"

	argsWithoutProg := os.Args[1:]
	// Default to dev config, if config file passed as arg, use it.
	if len(argsWithoutProg) > 0 {
		return argsWithoutProg[0]
	}
	return devConfigPath
}

func init() {
	configPath := getConfigPath()
	loadConfig(configPath)
}

func loadConfig(configPath string) {
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatalf("%v", err)
	}

	Configuration = configuration
}
