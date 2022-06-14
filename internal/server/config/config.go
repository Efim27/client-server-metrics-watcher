package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr string `env:"ADDRESS" envDefault:"127.0.0.1:8080"` //addr:port
	Store      struct {
		Interval string `env:"STORE_INTERVAL" envDefault:"300"`
		File     string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
		Restore  bool   `env:"RESTORE" envDefault:"true"`
	}
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
