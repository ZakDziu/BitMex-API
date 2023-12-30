package config

import (
	"github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload" // By design
)

type Configs struct {
	DBPostgresConfig DBPostgresConfig
	Server           ServerConfig
	Keys             Path
}

type DBPostgresConfig struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	User     string `env:"POSTGRES_USER"`
	DBName   string `env:"POSTGRES_DB"`
	Password string `env:"POSTGRES_PASSWORD"`
}

type Path struct {
	AccessKey  string `env:"HASH_KEY_ACCESS"`
	RefreshKey string `env:"HASH_KEY_REFRESH"`
}

type ServerConfig struct {
	ServerPort  string   `env:"SERVER_PORT"`
	ReadTimeout Duration `env:"READ_TIMEOUT"`
}

func New() (*Configs, error) {
	var config Configs
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
