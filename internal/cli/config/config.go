// Package config provides types and functions for working with configuration files used by the CLI.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type (
	// The Config type contains fields used to configure the CLI.
	Config struct {
		// The authentication token to use for API requests.
		Token string `toml:"token"`
	}
)

var (
	// ErrNotFound is the error given when a user's configuration file cannot be found.
	ErrNotFound = errors.New("not found")
)

// Load the configuration at the given path. Returns ErrNotFound if the file does not exist. This function expects
// configuration to be encoded as a TOML file.
func Load(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return config, ErrNotFound
	case err != nil:
		return config, err
	}

	return config, nil
}

// Save the configuration to the given path. Configuration is encoded as a TOML file.
func Save(path string, config Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(config)
}

// DefaultConfigPath returns the default path for storing a user's local configuration file.
func DefaultConfigPath() string {
	return filepath.Join(Dir(), "keeper", "config.toml")
}

// Dir returns a directory on-disk where configuration files/data can be stored. It handles different GOOS values and
// picks between the user's config and home directories.
func Dir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()

		switch runtime.GOOS {
		case "windows":
			dir = os.Getenv("APPDATA")
			if dir == "" {
				dir = filepath.Join(home, "AppData", "Roaming")
			}

		case "darwin":
			dir = filepath.Join(home, "Library", "Application Support")
		case "linux":
			dir = os.Getenv("XDG_CONFIG_HOME")
			if dir == "" {
				dir = filepath.Join(home, ".config")
			}
		default:
			dir = home
		}
	}

	return dir
}
