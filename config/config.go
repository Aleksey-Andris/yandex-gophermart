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
	HTTP
	Log
	PG
}

type HTTP struct {
	RunAddres      string `env:"RUN_ADDRESS"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

type Log struct {
	Environment string `env:"LOG_ENVEROMENT" envDefault:"develop"`
}

type PG struct {
	PoolMax int    `env:"POOL_MAX" envDefault:"1"`
	URI     string `env:"DATABASE_URI"`
	URL     string `env:"DATABASE_URL" envDefault:"postgres://postgres:gophermart@localhost:5432/postgres?sslmode=disable"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config error reading env: %w", err)
	}

	flag.StringVar(&flagRunAddress, "a", "", "run server address")
	flag.StringVar(&flagAccruralSystemAddresss, "r", "", "accrural system_address")
	flag.StringVar(&flagDataBaseURI, "d", "", "database uri")
	flag.Parse()

	if flagRunAddress != "" && cfg.HTTP.RunAddres == "" {
		cfg.HTTP.RunAddres = flagRunAddress
	}
	if flagAccruralSystemAddresss != "" && cfg.HTTP.AccrualAddress == "" {
		cfg.HTTP.AccrualAddress = flagAccruralSystemAddresss
	}
	if flagDataBaseURI != "" && cfg.PG.URI == "" {
		cfg.PG.URI = flagDataBaseURI
	}

	return cfg, nil
}
