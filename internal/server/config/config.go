package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// StoreConfig используется для хранения конфигурации агента, связанной с хранилищами.
type StoreConfig struct {
	// Interval - интервал выгрузки на диск (flag: i; default: 300s)
	Interval time.Duration `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	// DatabaseDSN - DSN БД (flag: d)
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	// File - файл для выгрузки (flag: f; default: /tmp/devops-metrics-db.json)
	File string `env:"STORE_FILE" json:"store_file,omitempty"`
	// Restore - чтение значений с диска при запуске (flag: r; default: false)
	Restore bool `env:"RESTORE" json:"restore,omitempty"`
}

// Config используется для хранения конфигурации сервера.
type Config struct {
	// ServerAddr - адрес сервера (flag: a; default: 127.0.0.1:8080)
	ServerAddr string `env:"ADDRESS" json:"address,omitempty"`
	// TrustedSubNet - строковое представление доверенной сети
	TrustedSubNet string `env:"TRUSTED_SUBNET" json:"trusted_subnet,omitempty"`
	// ProfilingAddr -  адрес WEB сервера профилировщика, не работает если пустое значение (flag: pa; default: 127.0.0.1:8090)
	ProfilingAddr string `env:"PROF_ADDRESS" json:"profiling_addr,omitempty"`
	// TemplatesAbsPath - абсолютный путь до шаблонов HTML (default: ./templates)
	TemplatesAbsPath string `env:"TEMPLATES_ABS_PATH" json:"templates_abs_path,omitempty"`
	// PrivateKeyRSA - приватный RSA ключ (flag: crypto-key)
	PrivateKeyRSA string `env:"CRYPTO_KEY" json:"crypto_key,omitempty"`
	// SignKey - ключ для подписи сообщений (flag: k)
	SignKey string `env:"KEY" json:"sign_key,omitempty"`
	// DebugMode - debug мод (flag: debug; default: false)
	DebugMode bool `env:"DEBUG" json:"debug,omitempty"`
	Store     StoreConfig
}

func newConfig() *Config {
	config := Config{}
	config.initDefaultValues()

	return &config
}

// initDefaultValues - значения конфига по умолчанию.
func (config *Config) initDefaultValues() {
	config.ServerAddr = "127.0.0.1:8080"
	config.TemplatesAbsPath = "./templates"
	config.Store = StoreConfig{
		Interval: time.Duration(300) * time.Second,
		File:     "/tmp/devops-metrics-db.json",
		Restore:  true,
	}
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
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		log.Println(err)
	}
}

func (config *Config) parseEnv() error {
	return env.Parse(config)
}

func (config *Config) parseFlags() {
	flag.StringVar(&config.ServerAddr, "a", config.ServerAddr, "server address (host:port)")
	flag.StringVar(&config.TrustedSubNet, "t", config.TrustedSubNet, "trusted subnet")
	flag.StringVar(&config.ProfilingAddr, "pa", config.ProfilingAddr, "profiling address (host:port)")
	flag.StringVar(&config.PrivateKeyRSA, "crypto-key", config.PrivateKeyRSA, "RSA private key")
	flag.StringVar(&config.SignKey, "k", config.SignKey, "sign key")
	flag.BoolVar(&config.DebugMode, "debug", config.DebugMode, "debug mode")

	//StoreConfig
	flag.BoolVar(&config.Store.Restore, "r", config.Store.Restore, "restoring metrics from file")
	flag.DurationVar(&config.Store.Interval, "i", config.Store.Interval, "store interval (example: 10s)")
	flag.StringVar(&config.Store.DatabaseDSN, "d", config.Store.DatabaseDSN, "Database DSN")
	flag.StringVar(&config.Store.File, "f", config.Store.File, "path to file for storage metrics")
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
