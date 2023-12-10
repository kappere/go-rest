package config

import (
	"fmt"

	"github.com/kappere/go-rest/core/config"
)

type Config struct {
	config.BaseConfig `yaml:",inline"`
}

func Load(path string) *Config {
	c := Config{
		BaseConfig: config.DefaultBaseConfig,
	}
	err := config.Load(path, &c)
	if err != nil {
		panic(fmt.Sprintf("Load config file failed! Path: %s, error: %v", path, err))
	}
	return &c
}
