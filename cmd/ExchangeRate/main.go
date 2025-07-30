package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sashaem1/ExchangeRate/internal/api"
	"github.com/sashaem1/ExchangeRate/internal/dataBase/postgresql"
	"github.com/sashaem1/ExchangeRate/internal/handler"
	"github.com/sashaem1/ExchangeRate/internal/server"

	_ "github.com/lib/pq"
)

func init() {
	if os.Getenv("ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Предупреждение: Не удалось найти .env файл")
		}
	}
}

func main() {
	handlers := handler.Handler{
		Api: api.NewFreeCurrencyApi(),
		DB:  postgresql.NewPostgreSqlDB(),
	}

	handlers.DB.InitDB(handlers.Api)
	handlers.DB.CronUpdateData(handlers.Api)
	defer handlers.DB.CloseConnect()

	srv := new(server.Server)
	err := srv.Run("8000", handlers.InitRouters())
	if err != nil {
		log.Fatalf("Ошибка старта сервера: %s", err)
	}
}
