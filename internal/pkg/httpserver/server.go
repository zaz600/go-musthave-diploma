package httpserver

import (
	"net/http"
	"time"
)

const (
	_defaultReadTimeout  = 5 * time.Second
	_defaultWriteTimeout = 5 * time.Second
	_defaultAddr         = ":8080"
)

func New(handler http.Handler, opts ...Option) *http.Server {
	s := &http.Server{
		Handler:      handler,
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
		Addr:         _defaultAddr,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
