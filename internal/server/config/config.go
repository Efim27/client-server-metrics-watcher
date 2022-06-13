package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr string `env:"ADDRESS" envDefault:"127.0.0.1:8080"` //addr:port
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
