package config

import "os"

type Config struct {
	Listen   string
	LogLevel string
}

func Load() *Config {
	var listen, logLevel string

	if listen = os.Getenv("LISTEN"); listen == "" {
		listen = "localhost:8888"
	}

	if logLevel = os.Getenv("LOGLEVEL"); logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		Listen:   listen,
		LogLevel: logLevel,
	}
}
