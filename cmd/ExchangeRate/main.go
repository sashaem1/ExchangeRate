package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashaem1/ExchangeRate/internal"
	"github.com/sashaem1/ExchangeRate/internal/api/http"
	freecurrencyapi "github.com/sashaem1/ExchangeRate/internal/freeCurrencyAPI"
	"github.com/sashaem1/ExchangeRate/internal/postgresql"

	_ "github.com/lib/pq"
)

func main() {
	pgxPool := initDbConnect()
	exchangeStorage := postgresql.NewExchangeStorage(pgxPool)
	externalAPIKey := os.Getenv("FREECURRENCY_API_KEY")
	ExchangeExternalAPI := freecurrencyapi.NewExchangeExternalAPI(externalAPIKey)
	exchangeRepo := internal.NewExchangeRepository(exchangeStorage, ExchangeExternalAPI)

	apiKeyStorage := postgresql.NewAPIKeyStorage(pgxPool)
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

func initDbConnect() *pgxpool.Pool {
	op := "main.main.initDbConnect"
	ctx := context.Background()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatalf("Не хватает данных из переменных окружения для подключения к бд")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("%s: %s", op, err)
	}

	const maxAttempts = 10
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {

		// Пытаемся выполнить пинг
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

// func main() {
// 	handlers := handler.Handler{
// 		Api: api.NewFreeCurrencyApi(),
// 		DB:  postgresql.NewPostgreSqlDB(),
// 	}

// 	handlers.DB.InitDB(handlers.Api)
// 	handlers.DB.CronUpdateData(handlers.Api)
// 	defer handlers.DB.CloseConnect()

// 	srv := new(server.Server)
// 	err := srv.Run("8000", handlers.InitRouters())
// 	if err != nil {
// 		log.Fatalf("Ошибка старта сервера: %s", err)
// 	}
// }
