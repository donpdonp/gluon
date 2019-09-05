package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type settings struct {
	Id           string
	Key          string
	AdminChannel string
}

var (
	Settings    = settings{}
	config_file = "config.json"
)

func LoadSettings() {
	_, err := os.Stat(config_file)
	if err != nil {
		Settings.Id = Snowflake()
		Settings.Key = Snowflake()

		settings_bytes, _ := json.Marshal(Settings)
		ioutil.WriteFile(config_file, settings_bytes, 0644)
	}

	b, err := ioutil.ReadFile(config_file)
	if err != nil {
		panic(err)
	}

	json.Unmarshal(b, &Settings)

}
