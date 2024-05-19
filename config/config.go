package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken   string
	TelegramMiniAppUrl string
	TgStorageBotToken  string
	TgStorageChatID    int64
	IsPgEnabled        bool
	PgHost             string
	PgPort             uint16
	PgDatabase         string
	PgUser             string
	PgPassword         string
}

func InitFromEnv() *Config {
	_ = godotenv.Load(".env")

	pgPort, _ := strconv.ParseUint(os.Getenv("PG_PORT"), 10, 16)
	pgPortUint16 := uint16(pgPort)

	chatID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_STORAGE_CHAT_ID"), 10, 64)

	return &Config{
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramMiniAppUrl: os.Getenv("TELEGRAM_MINI_APP_URL"),
		TgStorageBotToken:  os.Getenv("TELEGRAM_STORAGE_BOT_TOKEN"),
		TgStorageChatID:    chatID,
		IsPgEnabled:        os.Getenv("PG_HOST") != "",
		PgHost:             os.Getenv("PG_HOST"),
		PgPort:             pgPortUint16,
		PgDatabase:         os.Getenv("PG_DATABASE"),
		PgUser:             os.Getenv("PG_USER"),
		PgPassword:         os.Getenv("PG_PASSWORD"),
	}
}
