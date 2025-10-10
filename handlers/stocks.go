package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Stock struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}

type YahooQuote struct {
	Symbol             string  `json:"symbol"`
	ShortName          string  `json:"shortName"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

type YahooResponse struct {
	QuoteResponse struct {
		Result []YahooQuote `json:"result"`
	} `json:"quoteResponse"`
}

// GetStocks godoc
// @Summary Get Brazilian stocks information
// @Description Retrieve a list of Brazilian stocks from B3 market using Yahoo Finance API
// @Tags stocks
// @Accept  json
// @Produce  json
// @Success 200 {array} Stock
// @Router /stocks [get]
func GetStocks(c *gin.Context) {
	symbols := "PETR4.SA,VALE3.SA,ITUB4.SA,BBDC4.SA,WEGE3.SA,ABEV3.SA,MGLU3.SA,PETR3.SA,ITSA4.SA,B3SA3.SA"
	url := "https://query1.finance.yahoo.com/v7/finance/quote?symbols=" + symbols

	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from Yahoo Finance"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yahoo Finance API returned error"})
		return
	}

	var yahooResp YahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&yahooResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	var stocks []Stock
	for _, quote := range yahooResp.QuoteResponse.Result {
		stocks = append(stocks, Stock{
			Symbol: quote.Symbol,
			Name:   quote.ShortName,
			Price:  quote.RegularMarketPrice,
		})
	}

	c.JSON(http.StatusOK, stocks)
}
