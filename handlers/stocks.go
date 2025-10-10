package handlers

import (
	"context"
	"encoding/json"
	"go-br-finance-api/config"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Stock struct {
	Symbol string  `json:"stock"`
	Name   string  `json:"name"`
	Price  float64 `json:"close"`
}

type BrapiResponse struct {
	Stocks []Stock `json:"stocks"`
}

// GetStocks godoc
// @Summary Get Brazilian stocks information
// @Description Retrieve a list of all Brazilian stocks from B3 market using brapi.dev API
// @Tags stocks
// @Accept  json
// @Produce  json
// @Success 200 {array} Stock
// @Router /stocks [get]
func GetStocks(c *gin.Context) {
	ctx := context.Background()

	// Check cache
	if config.RedisClient != nil {
		cached, err := config.RedisClient.Get(ctx, "stocks").Result()
		if err == nil {
			var stocks []Stock
			json.Unmarshal([]byte(cached), &stocks)
			c.JSON(http.StatusOK, stocks)
			return
		}
	}

	url := "https://brapi.dev/api/quote/list?limit=1000&type=stock&token=" + os.Getenv("BRAPI_TOKEN")

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from brapi.dev"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "brapi.dev API returned error"})
		return
	}

	var brapiResp BrapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&brapiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	// Cache the data
	if config.RedisClient != nil {
		data, _ := json.Marshal(brapiResp.Stocks)
		config.RedisClient.Set(ctx, "stocks", data, 30*time.Minute)
	}

	c.JSON(http.StatusOK, brapiResp.Stocks)
}
