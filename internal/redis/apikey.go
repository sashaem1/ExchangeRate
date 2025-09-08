package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sashaem1/ExchangeRate/internal"
)

type APIKeyStorage struct {
	rdb *redis.Client
}

func NewAPIKeyStorage(rClient *redis.Client) *APIKeyStorage {
	return &APIKeyStorage{rdb: rClient}
}

func (as *APIKeyStorage) Get(ctx context.Context, APIKey string) (internal.APIKey, error) {
	op := "redis.apikey.Get"

	key := fmt.Sprintf("apikey:%s", APIKey)
	result := internal.NewAPIKey(APIKey)

	val, err := as.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("%s: %s", op, err)
	}

	// Если ключ есть, устанавливаем Valid=true
	result.Valid = val == "1"
	return result, nil
}

func (as *APIKeyStorage) Set(ctx context.Context, APIKey internal.APIKey) error {
	op := "redis.apikey.Set"

	key := fmt.Sprintf("apikey:%s", APIKey.Key)
	err := as.rdb.Set(ctx, key, "1", 0).Err()
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}

	return nil
}
