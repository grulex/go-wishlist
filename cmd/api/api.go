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
	// usecases:
	// — create a new user
	// — authenticate a user
	// — create a new wishlist
	// — subscribe to a wishlist
	// — unsubscribe from a wishlist
	// — add an item to a wishlist
	// — archive an item from a wishlist
	// — mark an item as booked
	// — mark an item as unbooked

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
