package config

import "os"

type Config struct {
	DbListen string
	Listen   string
	LogLevel string
}

func Load() *Config {
	var dbListen, listen, logLevel string

	if dbListen = os.Getenv("DB_LISTEN"); listen == "" {
		dbListen = "redis:6379"
	}
	if listen = os.Getenv("LISTEN"); listen == "" {
		listen = "localhost:8888"
	}

	if logLevel = os.Getenv("LOGLEVEL"); logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		DbListen: dbListen,
		Listen:   listen,
		LogLevel: logLevel,
	}
}
