package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Client struct {
	Server  int    `yaml:"ID"`
	Domain  string `yaml:"Domain"`
	Address string `yaml:"Address"`
}

type Config struct {
	Server struct {
		Host   string `yaml:"Host"`
		APIKey string `yaml:"APIKey"`
	} `yaml:"Server"`
	Clients []Client `yaml:"Clients"`
}

func LoadConfig(path string) (*Config, error) {
	var config *Config
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}
	return config, nil
}
