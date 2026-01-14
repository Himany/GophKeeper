package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	Host string `env:"SERVER_HOST" envDefault:"localhost"`
	Port int    `env:"SERVER_PORT" envDefault:"8080"`

	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://user:password@localhost/gophkeeper?sslmode=disable"`

	JWTSecret string        `env:"JWT_SECRET" envDefault:"your-secret-key"`
	JWTExpiry time.Duration `env:"JWT_EXPIRY" envDefault:"24h"`

	EncryptionKey string `env:"ENCRYPTION_KEY" envDefault:"12345678901234567890123456789012"`

	TLSCert string `env:"TLS_CERT_FILE"`
	TLSKey  string `env:"TLS_KEY_FILE"`

	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	LogFile  string `env:"LOG_FILE"`
}

type ClientConfig struct {
	ServerURL string `env:"SERVER_URL" envDefault:"http://localhost:8080"`

	DataDir string `env:"DATA_DIR" envDefault:".gophkeeper"`

	TokenFile string `env:"TOKEN_FILE" envDefault:".gophkeeper/token"`

	TLSSkipVerify bool   `env:"TLS_SKIP_VERIFY" envDefault:"false"`
	TLSCACert     string `env:"TLS_CA_CERT"`

	SyncInterval time.Duration `env:"SYNC_INTERVAL" envDefault:"5m"`

	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	LogFile  string `env:"LOG_FILE"`
}

func LoadServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadClientConfig() (*ClientConfig, error) {
	cfg := &ClientConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
