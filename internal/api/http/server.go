package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sashaem1/ExchangeRate/internal"
)

type ExchangeRepository interface {
	InitExchangeRepository(ctx context.Context) error
	GetByBase(baseCurrencyCode, targetCurrencyCode string) (internal.Exchange, error)
	GetByDate(date string) ([]internal.Exchange, error)
}

type APIKeyRepository interface {
	InitAPIKeyRepository(ctx context.Context) error
	VerificationAPIKey(apiKey string) (internal.APIKey, error)
}

type ActionLogRepository interface {
	InsertLog(ctx context.Context, ActionLog internal.ActionLog) error
}

type Server struct {
	httpServer          *http.Server
	exchangeRepository  ExchangeRepository
	apiKeyRepository    APIKeyRepository
	actionLogRepository ActionLogRepository
}

func NewServer(exchangeRepository ExchangeRepository, apiKeyRepository APIKeyRepository, actionLogRepository ActionLogRepository) *Server {
	return &Server{
		exchangeRepository:  exchangeRepository,
		apiKeyRepository:    apiKeyRepository,
		actionLogRepository: actionLogRepository,
	}
}

func (s *Server) Start(port string, handler http.Handler) error {
	op := "http.server.Start"
	ctx := context.Background()
	s.httpServer = &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
	}

	err := s.exchangeRepository.InitExchangeRepository(ctx)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	err = s.apiKeyRepository.InitAPIKeyRepository(ctx)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return s.httpServer.ListenAndServe()
}
