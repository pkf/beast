package global

import (
	"encoding/json"
	"os"
)

var Config = Configuration{}

func init() {
	Config, _ = LoadConfig("config/online.json")
	//Config = config
}

func LoadConfig(filename string) (Configuration, error) {
	var config = Configuration{}
	file, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}
