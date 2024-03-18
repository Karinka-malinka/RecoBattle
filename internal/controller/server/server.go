package server

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
	"golang.org/x/sync/errgroup"
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

func (s *Server) Start(ctx context.Context) {

	errs, _ := errgroup.WithContext(ctx)

	errs.Go(s.srv.ListenAndServe)
	log.Printf("server started: %s", s.srv.Addr)
	//logrus.Infof("server started: %s", s.srv.Addr)

	err := errs.Wait()
	if err != nil {
		log.Printf("message from server: %v", err)
	}
}

func (s *Server) Stop(ctx context.Context) {

	err := s.srv.Shutdown(ctx)
	if err != nil {
		log.Infof("server shutdown with error: %v", err)
	}
	log.Infof("Server is graceful shutdown...")
}
