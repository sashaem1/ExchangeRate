package internal

import (
	"context"
	"fmt"
	"os"
)

type APIKeyID string

type APIKey struct {
	ID    APIKeyID
	Key   string
	Valid bool
}

func NewAPIKey(key string) APIKey {
	return APIKey{Key: key}
}

type APIKeyStorage interface {
	Get(ctx context.Context, APIKey string) (APIKey, error)
	Set(ctx context.Context, APIKey APIKey) error
}

type APIKeyRepository struct {
	storage APIKeyStorage
}

func NewAPIKeyRepository(storage APIKeyStorage) *APIKeyRepository {
	return &APIKeyRepository{
		storage: storage,
	}
}

func (rr *APIKeyRepository) VerificationAPIKey(key string) (APIKey, error) {
	op := "internal.VerificationAPIKey"
	ctx := context.Background()

	verAPIKey, err := rr.storage.Get(ctx, key)
	if err != nil {
		return APIKey{}, fmt.Errorf("%s: %s", op, err)
	}

	return verAPIKey, nil
}

func (rr *APIKeyRepository) InitAPIKeyRepository(ctx context.Context) error {
	op := "internal.APIKey.initAPIKey"
	envAPIKeyStr := os.Getenv("DEFAULT_API_KEY")
	envAPIKey := NewAPIKey(envAPIKeyStr)

	err := rr.initAPIKey(ctx, envAPIKey)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}

func (rr *APIKeyRepository) initAPIKey(ctx context.Context, APIKey APIKey) error {
	op := "internal.initAPIKey"

	err := rr.storage.Set(ctx, APIKey)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
