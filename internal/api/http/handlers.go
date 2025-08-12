package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sashaem1/ExchangeRate/internal"
)

type Handler struct {
	exchangeRepository ExchangeRepository
}

type RateResponse struct {
	Base  string
	Rates map[string]float64 `json:"data"`
}

func NewHandler(exchangeRepository ExchangeRepository) *Handler {
	return &Handler{exchangeRepository: exchangeRepository}
}

func (h *Handler) InitRouters() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		rate := api.Group("/rate")
		{
			rate.GET("/current", h.getCurrentRateByBase)
			rate.GET("/historical", h.getCurrentRateByDate)
		}
	}

	return router
}

func (h *Handler) getCurrentRateByBase(c *gin.Context) {
	base := c.Query("base")
	symbol := c.Query("symbol")
	apiKey := c.Query("apikey")

	vAPI, err := h.exchangeRepository.VerificationAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !vAPI {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не верный API ключ"})
		return
	}

	exchange, err := h.exchangeRepository.GetByBase(base, symbol)
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
	date := c.Query("date")
	apiKey := c.Query("apikey")

	vAPI, err := h.exchangeRepository.VerificationAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !vAPI {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не верный API ключ"})
		return
	}

	exchanges, err := h.exchangeRepository.GetByDate(date)
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
