package config

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const DefaultEnvFile = ".env"

type ControlAPIConfig struct {
	DB_DSN             string `env:"CONTROL_DB_DSN,required,notEmpty"`
	Port               int    `env:"PORT" envDefault:"9090"`
	Mode               string `env:"MODE" envDefault:"release"`
	InternalAPIKey     string `env:"INTERNAL_API_KEY,required,notEmpty"`
	JWTSecret          string `env:"JWT_SECRET,required,notEmpty"`
	ProxyUrl           string `env:"PROXY_URL,required,notEmpty"`
	SuperadminUsername string `env:"SUPERADMIN_USERNAME,notEmpty"`
	SuperadminPassword string `env:"SUPERADMIN_PASSWORD,notEmpty"`
}

func NewControlAPIConfig() (*ControlAPIConfig, error) {
	cfg := &ControlAPIConfig{}
	if err := parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parse(cfg any) error {
	return env.Parse(cfg)
}

func LoadEnv(envFile string) error {
	if envFile != "" {
		return godotenv.Load(envFile)
	}

	if _, err := os.Stat(DefaultEnvFile); err == nil {
		return godotenv.Load(DefaultEnvFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
