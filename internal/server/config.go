package server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/davidsbond/keeper/internal/cli/config"
)

type (
	// The Config type contains all configuration values for the keeper server.
	Config struct {
		// Configuration for serving HTTP requests.
		HTTP HTTPConfig `toml:"http"`
		// Configuration for the master & account databases.
		Database DatabaseConfig `toml:"database"`
		// Configuration for generating/parsing JWT tokens used for authentication.
		JWT JWTConfig `toml:"jwt"`
	}

	// The HTTPConfig type contains fields used to configure the HTTP server.
	HTTPConfig struct {
		// The bind address of the HTTP server.
		Bind string `toml:"bind"`
	}

	// The DatabaseConfig type contains fields used to configure storage for the master and individual account
	// databases.
	DatabaseConfig struct {
		// The path for storing all databases.
		Path string `toml:"path"`
		// The TTL of individual account databases. Once opened, they will automatically be closed after this
		// time.
		TTL time.Duration `toml:"ttl"`
		// The encryption key for the master database.
		MasterKey string `toml:"master_key"`
	}

	// The JWTConfig type contains fields used to configure JWT tokens generated/parsed via the server.
	JWTConfig struct {
		// The JWT issuer.
		Issuer string `toml:"issuer"`
		// The TTL of each token.
		TTL time.Duration `toml:"ttl"`
		// The key used to sign and verify each JWT token.
		SigningKey string `toml:"signing_key"`
		// The JWT audience.
		Audience string `toml:"audience"`
	}
)

// DefaultConfig returns a Config instance with default configuration. This can be used for development/testing but
// ideally should not be used for a real deployment.
func DefaultConfig() Config {
	return Config{
		HTTP: HTTPConfig{
			Bind: "0.0.0.0:8080",
		},
		Database: DatabaseConfig{
			Path:      defaultDatabasePath(),
			TTL:       time.Hour,
			MasterKey: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0}, 32)),
		},
		JWT: JWTConfig{
			Issuer:     "dev",
			TTL:        time.Hour,
			SigningKey: base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0}, 32)),
			Audience:   "dev",
		},
	}
}

// LoadConfig attempts to parse a TOML file at the specified path into a Config type.
func LoadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate the Config.
func (c Config) Validate() error {
	return errors.Join(
		c.HTTP.validate(),
		c.Database.validate(),
		c.JWT.validate(),
	)
}

func (c HTTPConfig) validate() error {
	if _, _, err := net.SplitHostPort(c.Bind); err != nil {
		return fmt.Errorf("invalid http bind address: %q", c.Bind)
	}

	return nil
}

func (c DatabaseConfig) validate() error {
	if c.Path == "" {
		return errors.New("database path is required")
	}

	if c.TTL < time.Minute {
		return errors.New("database ttl must be at least 1 minute")
	}

	return nil
}

func (c JWTConfig) validate() error {
	if c.Issuer == "" {
		return errors.New("jwt issuer is required")
	}

	if c.TTL == 0 {
		return errors.New("jwt ttl is required")
	}

	if c.Audience == "" {
		return errors.New("jwt audience is required")
	}

	if len(c.SigningKey) == 0 {
		return errors.New("jwt signing key is required")
	}

	return nil
}

func defaultDatabasePath() string {
	return filepath.Join(config.Dir(), "keeper", "data")
}
