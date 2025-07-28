package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sashaem1/ExchangeRate/internal/api"
	database "github.com/sashaem1/ExchangeRate/internal/dataBase"
)

type Handler struct {
	Api api.ExchangeRates
	DB  database.DataBase
}

func (h *Handler) InitRouters() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		rate := api.Group("/rate")
		{
			rate.GET("/current", h.GetCurrentRate)
			rate.GET("/historical", h.GetHistoricalRate)
		}
	}

	return router
}
