package bot

import (
	"errors"
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	CommandPrefix string
	Operator      string
	Account       struct {
		Username string
		Password string
	}
	Database      struct {
		Host     string
		Port     int
		Username string
		Password string
		Database string
	}
	IRC           struct {
		Host string
		Port int
	}
}

func LoadConfiguration() (*Configuration, error) {
	content, err := ioutil.ReadFile("config.json")

	if err != nil {
		return nil, err
	}

	config := Configuration{}

	if json.Unmarshal(content, &config) != nil {
		return &config, errors.New("Could not load configuration.")
	}

	if len(config.Operator) == 0 {
		return &config, errors.New("You must configure an operator.")
	}

	return &config, nil
}
