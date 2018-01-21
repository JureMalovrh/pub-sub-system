package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type databaseConfig struct {
	Server     string
	Port       string
	Table      string
	Collection string
}

type publisherConfig struct {
	URL    string
	Port   string
	Method string
}

//Config definition
type Config struct {
	Address   string
	Database  databaseConfig
	Publisher publisherConfig
}

//LoadConfig loads config from path and returns loaded config
func LoadConfig(path string) *Config {
	defaultConfig := defaultConfig()
	fmt.Println(path)
	fmt.Println(defaultConfig)
	if _, err := toml.DecodeFile(path, defaultConfig); err != nil {
		log.Fatal("error", err.Error())
	}
	fmt.Println(defaultConfig)
	return defaultConfig
}

func defaultConfig() *Config {
	return &Config{
		Address: "localhost:8080",
		Database: databaseConfig{
			Server:     "localhost",
			Port:       "27017",
			Table:      "tracker",
			Collection: "user",
		},
		Publisher: publisherConfig{
			URL:    "localhost",
			Port:   "8000",
			Method: "ws",
		},
	}
}
