package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Pattern   string
	Whitelist []string
	Workers   int
	Verbose   bool
	ShowIPs   bool
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
	cfg := &Config{}
	if _, err := toml.DecodeFile(file, cfg); err != nil {
		return nil, err
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
