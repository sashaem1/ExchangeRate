package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sashaem1/ExchangeRate/internal"
)

type ExchangeRepository interface {
	InitExchangeRepository() error
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
	op := "http.Start"
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	err := s.exchangeRepository.InitExchangeRepository()
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}
	return s.httpServer.ListenAndServe()
}
