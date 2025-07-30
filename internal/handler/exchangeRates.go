package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetHistoricalRate(c *gin.Context) {
	date := c.Query("date")
	apiKey := c.Query("apikey")

	err := h.DB.InsertLog("pair")
	if err != nil {
		log.Println("Ошибка внесения лога запроса в бд: ", err)
	}

	verRes, err := h.DB.VerifyApiKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !verRes {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не подходящий апи-ключ"})
		return
	}

	if err := h.Api.ValidateDate(date); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Некорректный формат даты: %s", err.Error())})
		return
	}

	historicalRates, missingHistoricalRates, err := h.DB.GetRatesByDate(h.Api.GetDefaultBase(), date)
	log.Print(missingHistoricalRates)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(missingHistoricalRates) != 0 {
		ratesApi, err := h.DB.FillMissingData(date, missingHistoricalRates, h.Api)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"date":  date,
			"rates": ratesApi,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"date":  date,
			"rates": historicalRates,
		})
	}

}

func (h *Handler) GetCurrentRate(c *gin.Context) {
	base := c.Query("base")
	symbol := c.Query("symbol")
	apiKey := c.Query("apikey")

	err := h.DB.InsertLog("date")
	if err != nil {
		log.Println("Ошибка внесения лога запроса в бд: ", err)
	}

	verRes, err := h.DB.VerifyApiKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !verRes {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не подходящий апи-ключ"})
		return
	}

	if err := h.Api.ValidateRate(base); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Api.ValidateRate(symbol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pairRateDB, err := h.DB.GetRatesByPair(base, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if pairRateDB.Base == "" {
		responseRates, err := h.Api.GetLatestExchangeRatesByBase(base, symbol)

		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		h.DB.InsertRate(responseRates.Base, symbol, responseRates.Rates[symbol], time.Now().Format("2006-01-02"))

		c.JSON(http.StatusOK, gin.H{
			"base": responseRates.Base,
			"rate": responseRates.Rates,
		})

	} else {
		c.JSON(http.StatusOK, gin.H{
			"base": pairRateDB.Base,
			"rate": map[string]float64{
				pairRateDB.Currency: pairRateDB.Rate,
			},
		})
	}

}
