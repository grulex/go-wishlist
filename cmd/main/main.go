package main

import (
	"context"
	"fmt"
	"github.com/grulex/go-wishlist/bot"
	configPkg "github.com/grulex/go-wishlist/config"
	"github.com/grulex/go-wishlist/container"
	"github.com/grulex/go-wishlist/db"
	"github.com/grulex/go-wishlist/http"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	config := configPkg.InitFromEnv()
	if config.TelegramBotToken == "" {
		log.Fatal("env TELEGRAM_BOT_TOKEN is not set")
	}

	var serviceContainer *container.ServiceContainer
	if config.IsPgEnabled {
		dbConfig := db.Config{
			Host:     config.PgHost,
			Port:     config.PgPort,
			Database: config.PgDatabase,
			User:     config.PgUser,
			Password: config.PgPassword,
		}
		dbConnect, err := db.CreateDBConnection(dbConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer func(dbConnect *sqlx.DB) {
			_ = dbConnect.Close()
		}(dbConnect)
		serviceContainer = container.NewServiceContainer(dbConnect)
	} else {
		serviceContainer = container.NewInMemoryServiceContainer()
	}

	server := http.NewServer(":8080", serviceContainer, config)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered. Error:\n", r)
			}
		}()
		if err := server.Run(); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		b := bot.NewTelegramBot(config.TelegramBotToken, config.TelegramMiniAppUrl, serviceContainer)
		_ = b.Start()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("shutting down")
	os.Exit(0)
}
