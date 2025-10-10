package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go-br-finance-api/config"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	var allStocks []Stock
	if config.RedisClient != nil {
		cached, err := config.RedisClient.Get(ctx, "stocks").Result()
		if err == nil {
			json.Unmarshal([]byte(cached), &allStocks)
		}
	}

	// If not cached, fetch all
	if len(allStocks) == 0 {
		url := fmt.Sprintf("https://brapi.dev/api/quote/list?limit=1000&type=stock&token=%s", os.Getenv("BRAPI_TOKEN"))

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

		allStocks = brapiResp.Stocks

		// Cache the data
		if config.RedisClient != nil {
			data, _ := json.Marshal(allStocks)
			config.RedisClient.Set(ctx, "stocks", data, 30*time.Minute)
		}
	}

	// Apply search
	search := strings.ToLower(c.DefaultQuery("search", ""))
	var filteredStocks []Stock
	if search != "" {
		for _, stock := range allStocks {
			if strings.Contains(strings.ToLower(stock.Symbol), search) || strings.Contains(strings.ToLower(stock.Name), search) {
				filteredStocks = append(filteredStocks, stock)
			}
		}
	} else {
		filteredStocks = allStocks
	}

	// Apply limit
	limitStr := c.DefaultQuery("limit", "30")
	limit := 30
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	if limit > len(filteredStocks) {
		limit = len(filteredStocks)
	}

	stocks := filteredStocks[:limit]

	c.JSON(http.StatusOK, stocks)
}
