package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// HTTPClientConfig используется для хранения конфигурации агента, связанной с настройкой http клиента.
type HTTPClientConfig struct {
	// RetryCount - количество попыток отправки запроса (default: 2)
	RetryCount int `env:"RETRY_CONN_COUNT" json:"retry_count,omitempty"`
	// RetryWaitTime - время ожидания между попытками отправки запроса (default: 10s)
	RetryWaitTime time.Duration `env:"RETRY_CONN_WAIT_TIME"  json:"retry_wait_time,omitempty"`
	// RetryMaxWaitTime - макс. время для попыток отправки запроса (default: 90s)
	RetryMaxWaitTime time.Duration `env:"RETRY_CONN_MAX_WAIT_TIME"  json:"retry_max_wait_time,omitempty"`
	// ServerAddr - адрес сервера (default: 127.0.0.1:8080)
	ServerAddr string `env:"ADDRESS" json:"address,omitempty"`
	// isEnabledHTTPS - протокол HTTPS если true, иначе HTTP (default: true)
	IsEnabledHTTPS bool `env:"IS_ENABLED_HTTPS"  json:"is_enabled_https,omitempty"`
}

// Config используется для хранения конфигурации агента.
type Config struct {
	// PollInterval - интервал между считыванием метрик (flag: p; default: 2s)
	PollInterval time.Duration `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	// ReportInterval - интервал между отправки метрик (flag: r; default: 2s)
	ReportInterval time.Duration `env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
	// PublicKeyRSA - публичный RSA ключ (flag: crypto-key)
	PublicKeyRSA string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	// SignKey - ключ для подписи сообщений (flag: k)
	SignKey string `env:"KEY" json:"sign_key,omitempty"`
	// LogFile - лог файл (flag: l)
	LogFile string `env:"LOG_FILE" json:"log_file,omitempty"`
	// DebugMode - debug мод (flag: d)
	DebugMode            bool `env:"DEBUG" json:"debug,omitempty"`
	HTTPClientConnection HTTPClientConfig
}

// initDefaultValues - значения конфига по умолчанию.
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

func (config *Config) parseConfig(flagConfigPath, flagConfigPathAlias *string) {
	var configPath string
	if *flagConfigPath != "" {
		configPath = *flagConfigPath
	}

	if *flagConfigPathAlias != "" {
		configPath = *flagConfigPathAlias
	}

	if path, ok := os.LookupEnv("CONFIG"); ok {
		configPath = path
	}

	if configPath == "" {
		return
	}

	file, err := os.OpenFile(configPath, os.O_RDONLY|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
}

func (config *Config) parseEnv() error {
	return env.Parse(config)
}

func (config *Config) parseFlags() {
	flag.DurationVar(&config.ReportInterval, "r", config.ReportInterval, "report interval (example: 10s)")
	flag.DurationVar(&config.PollInterval, "p", config.PollInterval, "poll interval (example: 10s)")
	flag.StringVar(&config.HTTPClientConnection.ServerAddr, "a", config.HTTPClientConnection.ServerAddr, "server address (host:port)")
	flag.StringVar(&config.PublicKeyRSA, "crypto-key", config.PublicKeyRSA, "RSA public key")
	flag.StringVar(&config.SignKey, "k", config.SignKey, "sign key")
	flag.StringVar(&config.LogFile, "l", config.LogFile, "path to log file, to disable use empty path \"\"")
	flag.BoolVar(&config.DebugMode, "d", config.DebugMode, "debug mode")
	flag.Parse()
}

func LoadConfig() Config {
	config := newConfig()

	flagConfigPath := flag.String("c", "", "path to json config")
	flagConfigPathAlias := flag.String("config", "", "path to json config")

	config.parseFlags()
	config.parseConfig(flagConfigPath, flagConfigPathAlias)
	err := config.parseEnv()
	if err != nil {
		log.Fatal(err)
	}

	return *config
}
