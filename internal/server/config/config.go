package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

// StoreConfig используется для хранения конфигурации агента, связанной с хранилищами.
type StoreConfig struct {
	// Interval - интервал выгрузки на диск (flag: i; default: 300s)
	Interval time.Duration `env:"STORE_INTERVAL"`
	// DatabaseDSN - DSN БД (flag: d)
	DatabaseDSN string `env:"DATABASE_DSN"`
	// File - файл для выгрузки (flag: f; default: /tmp/devops-metrics-db.json)
	File string `env:"STORE_FILE"`
	// Restore - чтение значений с диска при запуске (flag: r; default: false)
	Restore bool `env:"RESTORE"`
}

// Config используется для хранения конфигурации сервера.
type Config struct {
	// ServerAddr - адрес сервера (flag: a; default: 127.0.0.1:8080)
	ServerAddr string `env:"ADDRESS"`
	// ProfilingAddr -  адрес WEB сервера профилировщика, не работает если пустое значение (flag: pa; default: 127.0.0.1:8090)
	ProfilingAddr string `env:"PROF_ADDRESS"`
	// TemplatesAbsPath - абсолютный путь до шаблонов HTML (default: ./templates)
	TemplatesAbsPath string `env:"TEMPLATES_ABS_PATH"`
	// SignKey - ключ для подписи сообщений (flag: k)
	SignKey string `env:"KEY"`
	// DebugMode - debug мод (flag: debug; default: false)
	DebugMode bool `env:"DEBUG"`
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

func (config *Config) parseEnv() error {
	return env.Parse(config)
}

func (config *Config) parseFlags() {
	flag.StringVar(&config.ServerAddr, "a", config.ServerAddr, "server address (host:port)")
	flag.StringVar(&config.ProfilingAddr, "pa", config.ProfilingAddr, "profiling address (host:port)")
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

	config.parseFlags()
	err := config.parseEnv()
	if err != nil {
		log.Fatal(err)
	}

	return *config
}
