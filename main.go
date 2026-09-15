package main

import (
	"context"
	"log"
	handler "wallet/handlers"
	"wallet/server"
	"wallet/services"
	"wallet/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	dsn := "postgres://postgres:Svs1512!@localhost:5432/testdb"
	strg, err := storage.NewDataBase(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer cancel()
		<-ctx.Done()

	}()
	srvc := services.NewWalett(strg)
	hndlrs := handler.CreateHandlers(srvc)
	srv := server.CreateNewServer(hndlrs)

	srv.StartServer()

}
