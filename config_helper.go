package tix

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ilyakaznacheev/cleanenv"
)

// DefaultConfig returns default config.
func DefaultConfig[ConfigT any]() *ConfigT {
	var cfg ConfigT
	// read configuration from the file and environment variables
	// or use default values if not set
	cleanenv.ReadEnv(&cfg)
	return &cfg
}

// LoadConfig loads config from file.
func LoadConfig[ConfigT any](path string) (*ConfigT, error) {
	var cfg ConfigT
	// read configuration from the file and environment variables
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, err
		} else {
			return &cfg, nil
		}
	}
	return &cfg, nil
}

// LoadConfigFromEnv loads config from environment variables.
func MustLoadConfig[ConfigT any](path string) *ConfigT {
	cfg, err := LoadConfig[ConfigT](path)
	if err != nil {
		panic(err)
	}
	return cfg
}

// PrintSampleConfig prints sample config to stdout.
func PrintSampleConfig[ConfigT any]() {
	cfg := DefaultConfig[ConfigT]()
	toml.NewEncoder(os.Stdout).Encode(cfg)
}
