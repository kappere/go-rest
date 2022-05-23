package config

import (
	"embed"

	"wataru.com/go-rest/core/rest"
)

type Config struct {
	rest.Config `yaml:"Config"`
}

var Conf Config

//go:embed *.yaml
var ConfFs embed.FS
