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
	PollInterval         time.Duration `env:"POLL_INTERVAL"`
	ReportInterval       time.Duration `env:"REPORT_INTERVAL"`
	SignKey              string        `env:"KEY"`
	LogFile              string        `env:"LOG_FILE"`
	DebugMode            bool          `env:"DEBUG"`
	HTTPClientConnection HTTPClientConfig
}

func (config *Config) initDefaultValues() {
	config.PollInterval = time.Duration(2) * time.Second
	config.ReportInterval = time.Duration(10) * time.Second
	config.DebugMode = false

	config.HTTPClientConnection = HTTPClientConfig{
		RetryCount:       2,
		RetryWaitTime:    time.Duration(10) * time.Second,
		RetryMaxWaitTime: time.Duration(90) * time.Second,
		ServerAddr:       "127.0.0.1:8080",
	}
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
	flag.DurationVar(&config.ReportInterval, "r", config.ReportInterval, "report interval (example: 10s)")
	flag.DurationVar(&config.PollInterval, "p", config.PollInterval, "poll interval (example: 10s)")
	flag.StringVar(&config.HTTPClientConnection.ServerAddr, "a", config.HTTPClientConnection.ServerAddr, "server address (host:port)")
	flag.StringVar(&config.SignKey, "k", config.SignKey, "sign key")
	flag.StringVar(&config.LogFile, "l", config.LogFile, "path to log file")
	flag.BoolVar(&config.DebugMode, "d", config.DebugMode, "debug mode")
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
