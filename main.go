package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"migrated-app/internal/resources"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv, err := resources.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("starting server on %s\n", addr)
	if listenErr := srv.ListenAddr(addr); listenErr != nil {
		select {
		case <-ctx.Done():
		default:
			log.Fatal(listenErr)
		}
	}
}