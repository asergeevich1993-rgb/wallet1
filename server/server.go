package server

import (
	"context"
	"errors"
	"net/http"
	"time"
	handler "wallet/handlers"
)

type Server struct {
	handler handler.HandlersInt
	svr     *http.Server
}

func CreateNewServer(hh handler.HandlersInt) *Server {

	router := http.NewServeMux()

	router.HandleFunc("POST /api/v1/wallets", hh.HandleCreateWallet)
	router.HandleFunc("POST /api/v1/wallet", hh.HandleChangeBalance)
	router.HandleFunc("GET /api/v1/wallets/{WALLET_UUID}", hh.HandleGetBalance)

	recover := Recover(router)

	return &Server{
		handler: hh,
		svr: &http.Server{
			Addr:              ":8080",
			Handler:           recover,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}
func (s *Server) StartServer() error {

	err := s.svr.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.svr.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
