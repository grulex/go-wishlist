package main

import (
	"context"
	configPkg "github.com/grulex/go-wishlist/config"
	"github.com/grulex/go-wishlist/container"
	"github.com/grulex/go-wishlist/http"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	config := configPkg.InitFromEnv()
	server := http.NewServer(":8080", container.NewInMemoryServiceContainer(), config)
	go func() {
		if err := server.Run(); err != nil {
			log.Println(err)
		}
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
