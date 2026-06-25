package config

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const DefaultEnvFile = ".env"

type ControlAPIConfig struct {
	DB_DSN         string   `env:"CONTROL_DB_DSN,required,notEmpty"`
	Port           int      `env:"PORT" envDefault:"9090"`
	Mode           string   `env:"MODE" envDefault:"release"`
	Debug          bool     `env:"DEBUG" envDefault:"false"`
	InternalAPIKey string   `env:"INTERNAL_API_KEY,required,notEmpty"`
	JWTSecret      string   `env:"JWT_SECRET,required,notEmpty"`
	ProxyUrl       string   `env:"PROXY_URL,required,notEmpty"`
	CORSOrigins    []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"*"`
	AppBaseURL     string   `env:"APP_BASE_URL" envDefault:"http://localhost:3000"`
	GoogleClientID string   `env:"GOOGLE_CLIENT_ID"`

	// SMTP delivers transactional email (welcome + password reset). When SMTPHost
	// is empty, email is disabled: reset links are logged and returned in the API
	// response instead. STARTTLS on port 587 is the supported transport.
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPFrom     string `env:"SMTP_FROM"`
	SMTPFromName string `env:"SMTP_FROM_NAME" envDefault:"xpaywall"`
}

// MailEnabled reports whether SMTP delivery is configured.
func (c *ControlAPIConfig) MailEnabled() bool {
	return c.SMTPHost != ""
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
