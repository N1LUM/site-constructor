package configs

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(addr string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func BuildAppAddress() string {
	return fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT"))
}
