package config

import (
	"embed"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	loadFromBytes(data, v)
	return nil
}

func LoadEmbed(configFs embed.FS, v any) error {
	return nil
}

func loadFromBytes(data []byte, v any) error {
	yaml.Unmarshal(data, v)
	return nil
}
