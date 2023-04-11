package safeconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Domain struct {
	Name string `yaml:"name"`
	Host string `yaml:"host,omitempty"`
}

type domainAlias Domain

func (a *Domain) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var d domainAlias
	if err := unmarshal(&d); err == nil {
		*a = Domain(d)
		return nil
	}

	var ds string
	if err := unmarshal(&ds); err != nil {
		return err
	}
	*a = Domain{Name: ds}
	return nil
}

type SafeConfig struct {
	Domains []Domain `yaml:"domains"`
}

func New(pathToFile string) (SafeConfig, error) {
	cfg := SafeConfig{}
	if pathToFile == "" {
		log.Debug().Msg("config file path is empty, skip loading")
		return cfg, nil
	}

	if err := cfg.Reload(pathToFile); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (cfg *SafeConfig) Reload(pathToFile string) error {
	log.Info().Msgf("trying to load config from file %s", pathToFile)

	filename, err := filepath.Abs(pathToFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of file %s: %w", pathToFile, err)
	}
	log.Debug().Msgf("absolute path of config file is %s", filename)

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	err = yaml.Unmarshal(yamlFile, cfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file: %w", err)
	}

	log.Debug().Msgf("config file is loaded:\n %s", *cfg)
	return nil
}
