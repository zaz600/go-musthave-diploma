package httpserver

import (
	"net/http"
	"time"
)

type Option func(*http.Server)

func WithAddr(addr string) Option {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *http.Server) {
		s.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *http.Server) {
		s.WriteTimeout = timeout
	}
}
