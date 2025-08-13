package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashaem1/ExchangeRate/internal"
)

type RateResponse struct {
	Base  string
	Rates map[string]float64 `json:"data"`
}

type Handler struct {
	server *Server
}

func NewHandler(server *Server) *Handler {
	return &Handler{server: server}
}

func (h *Handler) InitRouters() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		rate := api.Group("/rate")
		{
			rate.GET("/current", h.getCurrentRateByPair)
			rate.GET("/historical", h.getCurrentRateByDate)
		}
	}

	return router
}

func (h *Handler) getCurrentRateByPair(c *gin.Context) {
	op := "http.handlers.getCurrentRateByBase"
	base := c.Query("base")
	symbol := c.Query("symbol")
	apiKeyString := c.Query("apikey")
	ctx := context.Background()

	actionLog, err := internal.NewActionLog("pair", time.Now())
	if err != nil {
		log.Printf("%s: %s", op, err)
	}

	err = h.server.actionLogRepository.InsertLog(ctx, actionLog)
	if err != nil {
		log.Printf("%s: %s", op, err)
	}

	apiKey, err := h.server.apiKeyRepository.VerificationAPIKey(apiKeyString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !apiKey.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не верный API ключ"})
		return
	}

	exchange, err := h.server.exchangeRepository.GetByBase(base, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"base": exchange.BaseCurrency.Code,
		"rate": map[string]float64{
			exchange.TargetCurrency.Code: exchange.Rate,
		},
	})
}

func (h *Handler) getCurrentRateByDate(c *gin.Context) {
	op := "http.handlers.getCurrentRateByDate"
	date := c.Query("date")
	apiKeyString := c.Query("apikey")
	ctx := context.Background()

	actionLog, err := internal.NewActionLog("date", time.Now())
	if err != nil {
		log.Printf("%s: %s", op, err)
	} else {
		err = h.server.actionLogRepository.InsertLog(ctx, actionLog)
		if err != nil {
			log.Printf("%s: %s", op, err)
		}
	}

	apiKey, err := h.server.apiKeyRepository.VerificationAPIKey(apiKeyString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !apiKey.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не верный API ключ"})
		return
	}

	exchanges, err := h.server.exchangeRepository.GetByDate(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rates := ConvertExchangesToRateResponse(exchanges)

	c.JSON(http.StatusOK, gin.H{
		"date":  date,
		"rates": rates,
	})
}

func ConvertExchangesToRateResponse(exchanges []internal.Exchange) []RateResponse {
	rateMap := make(map[string]map[string]float64)

	for _, ex := range exchanges {
		base := ex.BaseCurrency.Code
		target := ex.TargetCurrency.Code

		if _, exists := rateMap[base]; !exists {
			rateMap[base] = make(map[string]float64)
		}

		rateMap[base][target] = ex.Rate
	}

	var result []RateResponse
	for base, rates := range rateMap {
		result = append(result, RateResponse{
			Base:  base,
			Rates: rates,
		})
	}

	return result
}
