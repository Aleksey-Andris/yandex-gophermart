package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v9"
)

var (
	flagRunAddress             string
	flagAccruralSystemAddresss string
	flagDataBaseURI            string
)

type Config struct {
	RunAddres      string `env:"RUN_ADDRESS"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogEnvironment string `env:"LOG_ENVEROMENT" envDefault:"develop"`
	DBPoolMax      int    `env:"POOL_MAX" envDefault:"1"`
	DBURI          string `env:"DATABASE_URI"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed reading env: %w", err)
	}

	flag.StringVar(&flagRunAddress, "a", "", "run server address")
	flag.StringVar(&flagAccruralSystemAddresss, "r", "", "accrural system_address")
	flag.StringVar(&flagDataBaseURI, "d", "", "database uri")
	flag.Parse()

	if flagRunAddress != "" && cfg.RunAddres == "" {
		cfg.RunAddres = flagRunAddress
	}
	if flagAccruralSystemAddresss != "" && cfg.AccrualAddress == "" {
		cfg.AccrualAddress = flagAccruralSystemAddresss
	}
	if flagDataBaseURI != "" && cfg.DBURI == "" {
		cfg.DBURI = flagDataBaseURI
	}

	return cfg, nil
}
