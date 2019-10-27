package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	amazonConfig "multiparser/provider/amazon/config"
	"os"
)

type Config struct {
	Common struct {
		Iplist    string `yaml:"iplist"`
		PackLimit int    `yaml:"packlimit"`
	}
	Providers struct {
		Amazon amazonConfig.Amazon
	}
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
