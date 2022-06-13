package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	HTTPClientConnection struct {
		RetryCount       int           `env:"RETRYCONNCOUNT" envDefault:"1"`
		RetryWaitTime    time.Duration `env:"RETRYCONNWAITTIME" envDefault:"10s"`
		RetryMaxWaitTime time.Duration `env:"RETRYCONNMAXWAITTIME" envDefault:"90s"`
	}
	PollInterval   time.Duration `env:"POLLINTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORTINTERVAL" envDefault:"10s"`
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
