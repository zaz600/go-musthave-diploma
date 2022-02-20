package config

import (
	"flag"
)

const (
	defaultServerAddress = "localhost:8080"
)

type AppConfig struct {
	ServerAddress  string
	DatabaseDSN    string
	AccrualAddress string
}

type RepoType string

const (
	MemoryRepo   RepoType = "Memory"
	DatabaseRepo RepoType = "DB"
)

func (s AppConfig) GetRepositoryType() RepoType {
	if s.DatabaseDSN != "" {
		return DatabaseRepo
	}
	return MemoryRepo
}

// GetConfig возвращает конфигурацию приложения, вычитывая в таком порядке
// аргументы командной строки -> env
// args - пока не используется
func GetConfig(args []string) *AppConfig {
	cfg := &AppConfig{}
	flag.StringVar(&cfg.ServerAddress, "a", getEnvOrDefault("RUN_ADDRESS", defaultServerAddress), "listen address. env: RUN_ADDRESS")
	flag.StringVar(&cfg.DatabaseDSN, "d", getEnvOrDefault("DATABASE_URI", ""), "PG dsn. env: DATABASE_URI")
	flag.StringVar(&cfg.AccrualAddress, "r", getEnvOrDefault("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual address. env: ACCRUAL_SYSTEM_ADDRESS")
	flag.Parse()
	return cfg
}
