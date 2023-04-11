package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Pattern   string   `toml:"pattern"`
	Whitelist []string `toml:"whitelist"`
	Workers   int      `toml:"workers"`
	Verbose   bool     `toml:"verbose"`
	ShowIPs   bool     `toml:"showips"`
}

func NewWithDefaults() *Config {
	return &Config{
		Workers: 16,
		Verbose: false,
		ShowIPs: false,
	}
}

func FindConfig() (string, error) {
	configDir, _ := os.UserConfigDir()
	homeDir, _ := os.UserHomeDir()

	defaultConfigFile := []string{
		"./.domaininator.toml",
		"./domaininator.toml",
		fmt.Sprintf("%s/domaininator/config.toml", configDir),
		fmt.Sprintf("%s/.domaininator.toml", homeDir),
		"/etc/domaininator.toml",
	}

	for _, f := range defaultConfigFile {
		info, err := os.Stat(f)
		if !os.IsNotExist(err) && !info.IsDir() {
			return f, nil
		}
	}
	return "", errors.New("No default config located")
}

func NewFromTOML(file string) (*Config, error) {
	cfg := NewWithDefaults()
	if _, err := toml.DecodeFile(file, cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (c *Config) InWhitelist(domain string) bool {
	for _, d := range c.Whitelist {
		if domain == d {
			return true
		}
	}
	return false
}
