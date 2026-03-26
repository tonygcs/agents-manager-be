package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Workerd struct {
		Addr string `yaml:"addr"`
	} `yaml:"workerd"`
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	GitHub struct {
		Token string `yaml:"token"`
	} `yaml:"github"`
}

func Load(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	return cfg, yaml.NewDecoder(f).Decode(&cfg)
}
