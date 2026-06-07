// Package config handles the configs
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var DefaultConfigFile = "conf.json"

var DefaultConfig = Config{
	Hosts: []HostConfig{
		{
			Host:     "localhost",
			Name:     "postgres",
			Port:     "5432",
			User:     "treemap_collector",
			Password: "verysecretpw",
		},
	},
	ServeAddr: "localhost:5432",
}

type Config struct {
	Hosts     []HostConfig `json:"hosts"`
	ServeAddr string       `json:"serve_addr"`
}

type HostConfig struct {
	Host     string `json:"host"`
	Name     string `json:"name"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func LoadConfig(args ...string) (Config, error) {
	configFile := ""
	if len(args) > 0 {
		configFile = args[0]
	} else {
		return Config{}, fmt.Errorf("config file was not provided")
	}

	var cfg Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func CreateConfigFile() error {
	data, err := json.MarshalIndent(DefaultConfig, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(
		"conf.json",
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o644,
	)
	if err != nil {
		// if file already exists
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}
