package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	handler "wallet/handlers"
	"wallet/server"
	"wallet/services"
	"wallet/storage"
)

func dsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	strdsn := dsn()
	strg, err := storage.NewDataBase(ctx, strdsn)
	if err != nil {
		log.Fatal(err)
	}

	shutdownDone := make(chan struct{})

	srvc := services.NewWallet(strg)
	hndlrs := handler.CreateHandlers(srvc)
	srv := server.CreateNewServer(hndlrs)
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		nctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(nctx)
		if err != nil {
			log.Printf("shutdown error %v", err)
		}
	}()

	err = srv.StartServer()
	if err != nil {
		log.Fatalf("error %v", err)
	}
	<-shutdownDone

	strg.Close()
}
