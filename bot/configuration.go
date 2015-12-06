package bot

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	CommandPrefix string `yaml:"commandPrefix"`
	Operator      string
	Account       struct {
		Username string
		Password string
	}
	Database struct {
		DSN string `yaml:"DSN"`
	}
	IRC struct {
		Host string
		Port int
	}
	Plugins map[string]interface{}
}

func LoadConfiguration(filename string) (*Configuration, error) {
	content, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	config := Configuration{}

	if yaml.Unmarshal(content, &config) != nil {
		return &config, errors.New("Could not load configuration file '" + filename + "'.")
	}

	if len(config.Operator) == 0 {
		return &config, errors.New("You must configure an operator.")
	}

	return &config, nil
}

func (self *Configuration) PluginConfig(plugin string, dest interface{}) error {
	data, exists := self.Plugins[plugin]

	if exists {
		// very lazy hack because i could not figure out how to nicely type assert
		// the existing structure (which seems to be an endless map[string]interface{}
		// monster) to the concrete dest struct
		encoded, _ := yaml.Marshal(data)

		err := yaml.Unmarshal(encoded, dest)
		if err != nil {
			return err
		}
	}

	return nil
}
