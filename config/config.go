package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	CommandList string `yaml:"commandlist"`
	PackLimit   int    `yaml:"packlimit"`
}

func (c *Config) Load(fn string) error {
	f, err := os.Open("./config.yaml")
	if err != nil {
		return fmt.Errorf("Load: %w ", err)
	}
	if err := yaml.NewDecoder(f).Decode(c); err != nil {
		return fmt.Errorf("Load: %w ", err)
	}
	return nil
}
