package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
)

type Server struct {
	srv http.Server
}

func NewServer(addr string, h http.Handler) *Server {

	s := &Server{}

	s.srv = http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Start() {

	go func() {
		err := s.srv.ListenAndServe()
		log.Printf("error server started: %v", err)
	}()

	log.Printf("server started: %s", s.srv.Addr)
}

func (s *Server) Stop(ctx context.Context) {

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*2))
	defer cancel()

	err := s.srv.Shutdown(timeoutCtx)
	if err != nil {
		log.Infof("server shutdown with error: %v", err)
	}
	log.Infof("Server is graceful shutdown...")
}
