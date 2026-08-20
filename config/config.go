package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	LogLevel string `json:"log_level"`
	Token    string `json:"token"`
	DataPath string `json:"data_path"`
}

func NewConfig() *Config {
	return &Config{
		LogLevel: "info",
		Token:    "",
		DataPath: "./data.bson",
	}
}

func (c *Config) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return err
	}
	return nil
}
