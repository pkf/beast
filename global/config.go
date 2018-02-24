package global

import (
	_ "beast/util"
	"encoding/json"
	_ "fmt"
	"os"
)

var Config = Configuration{}

func init() {
	//dir:= util.GetCurrentPath()
	//Config, _ = LoadConfig(dir+ "config/online.json")
}

func LoadConfig(filename string) (Configuration, error) {
	var config = Configuration{}
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
		return config, err
	}
	return config, nil
}
