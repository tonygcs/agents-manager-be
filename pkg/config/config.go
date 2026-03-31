package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type WorkerConfig struct {
	Image       string            `yaml:"image"`
	Description string            `yaml:"description"`
	Cmd         []string          `yaml:"cmd"`
	Labels      map[string]string `yaml:"labels"`
	Secrets     []string          `yaml:"secrets"`
}

type Config struct {
	Workerd struct {
		Addr string `yaml:"addr"`
	} `yaml:"workerd"`
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	Secrets map[string]string       `yaml:"secrets"`
	Workers map[string]WorkerConfig `yaml:"workers"`
}

func (c Config) validate() error {
	// Secrets.
	for workerName, workerCfg := range c.Workers {
		for _, key := range workerCfg.Secrets {
			if _, ok := c.Secrets[key]; !ok {
				return fmt.Errorf("missing secret %q required by worker %q", key, workerName)
			}
		}
	}
	return nil
}

func Load(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, cfg.validate()
}
