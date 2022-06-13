package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	HTTPClientConnection struct {
		RetryCount       int           `env:"RETRY_CONN_COUNT" envDefault:"2"`
		RetryWaitTime    time.Duration `env:"RETRY_CONN_WAIT_TIME" envDefault:"10s"`
		RetryMaxWaitTime time.Duration `env:"RETRY_CONN_MAX_WAIT_TIME" envDefault:"90s"`
	}
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	ServerAddr     string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"` //addr:port
}

func LoadConfig() Config {
	var config Config
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

var AppConfig Config = LoadConfig()
