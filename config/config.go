package config

import (
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken   string
	TelegramMiniAppUrl string
	IsPgEnabled        bool
	PgHost             string
	PgPort             uint16
	PgDatabase         string
	PgUser             string
	PgPassword         string
}

func InitFromEnv() *Config {
	pgPort, _ := strconv.ParseUint(os.Getenv("PG_PORT"), 10, 16)
	pgPortUint16 := uint16(pgPort)
	return &Config{
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramMiniAppUrl: os.Getenv("TELEGRAM_MINI_APP_URL"),
		IsPgEnabled:        os.Getenv("PG_HOST") != "",
		PgHost:             os.Getenv("PG_HOST"),
		PgPort:             pgPortUint16,
		PgDatabase:         os.Getenv("PG_DATABASE"),
		PgUser:             os.Getenv("PG_USER"),
		PgPassword:         os.Getenv("PG_PASSWORD"),
	}
}
