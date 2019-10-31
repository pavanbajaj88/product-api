package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// IConfig is an interface defining how to load configuration
type IConfig interface {
	LoadConfigs(string) error
}

// AWS contains configuration for the AWS DynamoDB
type AWS struct {
	AccessKeyID       string `json:"accessKeyId"`
	SecretAccessKeyID string `json:"secretAccessKeyId"`
	Region            string `json:"region"`
}

// Router contains configuration for the mux router
type Router struct {
	Port string `json:"port"`
}

type settings struct {
	AWS    `json:"aws"`
	Router `json:"router"`
}

// Settings contains the global configuration for the application
var Settings settings

// LoadConfigs loads configuration settings from JSON
func (s settings) LoadConfigs(file string) error {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&Settings)

	// no error, so return nil
	return nil
}