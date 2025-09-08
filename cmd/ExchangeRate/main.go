package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sashaem1/ExchangeRate/internal"
	"github.com/sashaem1/ExchangeRate/internal/api/http"
	freecurrencyapi "github.com/sashaem1/ExchangeRate/internal/freeCurrencyAPI"
	"github.com/sashaem1/ExchangeRate/internal/postgresql"
	internalRedis "github.com/sashaem1/ExchangeRate/internal/redis"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	pgxPool := initDbConnect(ctx, dbHost, dbPort, dbUser, dbName, dbPassword)
	exchangeStorage := postgresql.NewExchangeStorage(pgxPool)
	externalAPIKey := os.Getenv("FREECURRENCY_API_KEY")
	ExchangeExternalAPI := freecurrencyapi.NewExchangeExternalAPI(externalAPIKey)
	exchangeRepo := internal.NewExchangeRepository(exchangeStorage, ExchangeExternalAPI)

	redisAddr := os.Getenv("REDIS_ADDRES")
	rClient := initCacheConnect(ctx, redisAddr)
	apiKeyStorage := internalRedis.NewAPIKeyStorage(rClient)
	apiKeyRepo := internal.NewAPIKeyRepository(apiKeyStorage)

	actionLogStorage := postgresql.NewActionLogStorage(pgxPool)
	actionLogRepository := internal.NewActionLogRepository(actionLogStorage)

	httpServer := http.NewServer(exchangeRepo, apiKeyRepo, actionLogRepository)
	httpHandler := http.NewHandler(httpServer)

	err := httpServer.Start("8000", httpHandler.InitRouters())
	if err != nil {
		log.Fatalf("Ошибка старта сервера: %s", err)
	}

}

func initDbConnect(ctx context.Context, dbHost string, dbPort string, dbUser string, dbName string, dbPassword string) *pgxpool.Pool {
	op := "main.main.initDbConnect"

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatalf("Не хватает данных из переменных окружения для подключения к бд")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("%s: %s", op, err)
	}

	maxAttempts := 10
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {

		err = pgxPool.Ping(ctx)
		if err == nil {
			log.Printf("Успешное подключение к базе данных")
			return pgxPool
		}

		log.Printf("Пинг базы данных %d/%d. Ошибка: %v", attempt, maxAttempts, err)

		if attempt < maxAttempts {
			time.Sleep(retryDelay)
		}
	}

	pgxPool.Close()
	log.Fatalf("%s: %s", op, "Не удалось подключиться к бд")
	return nil
}

func initCacheConnect(ctx context.Context, redisAddr string) *redis.Client {
	op := "main.main.initCacheConnect"

	rClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	maxAttempts := 10
	retryDelay := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {

		_, err := rClient.Ping(ctx).Result()
		if err == nil {
			log.Printf("Успешное подключение к Redis")
			return rClient
		}

		log.Printf("Пинг Redis %d/%d. Ошибка: %v", attempt, maxAttempts, err)

		if attempt < maxAttempts {
			time.Sleep(retryDelay)
		}
	}

	log.Fatalf("%s: %s", op, "Не удалось подключиться к Redis")

	return rClient
}
