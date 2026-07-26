package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

	_ = srv.ListenAddr()

	fmt.Printf("starting server on %s\n", addr)
	httpSrv := &http.Server{Addr: addr}
	go func() {
		<-ctx.Done()
		httpSrv.Shutdown(context.Background())
	}()
	if listenErr := httpSrv.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
		log.Fatal(listenErr)
	}
}