package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Stock struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
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
	url := "https://brapi.dev/api/quote/list?token=" + os.Getenv("BRAPI_TOKEN")

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

	c.JSON(http.StatusOK, brapiResp.Stocks)
}
