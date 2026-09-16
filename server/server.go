package server

import (
	"context"
	"log"
	"net/http"
	handler "wallet/handlers"
)

type Server struct {
	handler handler.HandlersInt
	svr     *http.Server
}

func CreateNewServer(hh handler.HandlersInt) *Server {
	return &Server{
		handler: hh,
	}
}
func (s *Server) StartServer() {
	router := http.NewServeMux()

	router.HandleFunc("POST /api/v1/wallets", s.handler.HandleCreateWallet)
	router.HandleFunc("POST /api/v1/wallet", s.handler.HandleChangeBalance)
	router.HandleFunc("GET /api/v1/wallets/{WALLET_UUID}", s.handler.HandleGetBalance)

	saver := Saver(router)

	s.svr = &http.Server{
		Addr:    ":8080",
		Handler: saver,
	}
	if err := s.svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.svr.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
