package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type HTTPClientConfig struct {
	RetryCount       int           `env:"RETRY_CONN_COUNT"`
	RetryWaitTime    time.Duration `env:"RETRY_CONN_WAIT_TIME"`
	RetryMaxWaitTime time.Duration `env:"RETRY_CONN_MAX_WAIT_TIME"`
	ServerAddr       string        `env:"ADDRESS"` //addr:port
}

type Config struct {
	HTTPClientConnection HTTPClientConfig
	PollInterval         time.Duration `env:"POLL_INTERVAL"`
	ReportInterval       time.Duration `env:"REPORT_INTERVAL"`
}

func (config *Config) initDefaultValues() {
	config.HTTPClientConnection = HTTPClientConfig{
		RetryCount:       2,
		RetryWaitTime:    time.Duration(10) * time.Second,
		RetryMaxWaitTime: time.Duration(90) * time.Second,
		ServerAddr:       "127.0.0.1:8080",
	}

	config.PollInterval = time.Duration(2) * time.Second
	config.ReportInterval = time.Duration(10) * time.Second
}

func newConfig() *Config {
	config := Config{}
	config.initDefaultValues()

	return &config
}

func (config *Config) parseEnv() error {
	return env.Parse(config)
}

func (config *Config) parseFlags() {
	flag.StringVar(&config.HTTPClientConnection.ServerAddr, "a", config.HTTPClientConnection.ServerAddr, "server address (host:port)")
	flag.DurationVar(&config.ReportInterval, "r", config.ReportInterval, "report interval (example: 10s)")
	flag.DurationVar(&config.PollInterval, "p", config.PollInterval, "poll interval (example: 10s)")
	flag.Parse()
}

func LoadConfig() Config {
	config := newConfig()

	config.parseFlags()
	err := config.parseEnv()
	if err != nil {
		log.Fatal(err)
	}

	return *config
}
