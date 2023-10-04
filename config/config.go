package config

import "os"

type Config struct {
	TelegramBotToken string
}

func InitFromEnv() *Config {
	return &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
	}
}
