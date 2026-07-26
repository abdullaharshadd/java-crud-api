package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"migrated-app/internal/resources"
	"migrated-app/internal/smartcontact"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	cfg, err := resources.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", cfg.BuildDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	srv := &http.Server{
		Addr:    addr,
		Handler: smartcontact.BuildRouterWithDB(db),
	}

	go func() {
		fmt.Printf("starting server on %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
