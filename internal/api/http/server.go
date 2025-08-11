package http

import (
	"net/http"
	"time"

	"github.com/sashaem1/ExchangeRate/internal"
)

type ExchangeRepository interface {
	GetByBase(baseCurrencyCode, targetCurrencyCode string) (internal.Exchange, error)
	GetByDate(date string) ([]internal.Exchange, error)
	VerificationAPIKey(apiKey string) (bool, error)
}

type Server struct {
	httpServer         *http.Server
	exchangeRepository ExchangeRepository
}

func NewServer(exchangeRepository ExchangeRepository) *Server {
	return &Server{exchangeRepository: exchangeRepository}
}

func (s *Server) Start(port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}
