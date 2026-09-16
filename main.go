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

func main() {
	pctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := "postgres://postgres:Svs1512!@localhost:5432/testdb"
	strg, err := storage.NewDataBase(pctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	pid := os.Getpid()
	fmt.Println(pid)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM)
	go func() {
		<-sigchan
		cancel()
	}()

	srvc := services.NewWalett(strg)
	hndlrs := handler.CreateHandlers(srvc)
	srv := server.CreateNewServer(hndlrs)
	go func() {
		<-pctx.Done()
		sctx, scancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer scancel()
		srv.Shutdown(sctx)
	}()

	srv.StartServer()

}
