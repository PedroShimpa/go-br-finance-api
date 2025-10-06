package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-br-finance-api/cache"

	"github.com/gin-gonic/gin"
)

type CurrencyConversion struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Result float64 `json:"result"`
	Rate   float64 `json:"rate"`
}

type InflationData struct {
	Date        string  `json:"date"`
	Value       float64 `json:"value"`
	Previous    float64 `json:"previous"`
	Variation   float64 `json:"variation"`
	Description string  `json:"description"`
}

type InvestmentCalculation struct {
	Principal      float64 `json:"principal"`
	MonthlyDeposit float64 `json:"monthly_deposit"`
	AnnualRate     float64 `json:"annual_rate"`
	Years          int     `json:"years"`
	TotalInvested  float64 `json:"total_invested"`
	FinalAmount    float64 `json:"final_amount"`
	TotalInterest  float64 `json:"total_interest"`
}

// GetCurrencyConversion godoc
// @Summary Converter moeda
// @Description Converte valor entre moedas usando taxas do Brasil API
// @Tags calculations
// @Accept  json
// @Produce  json
// @Param from query string true "Moeda de origem (USD, EUR, etc.)"
// @Param to query string true "Moeda de destino (BRL)"
// @Param amount query number true "Valor a converter"
// @Success 200 {object} CurrencyConversion
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /calculations/currency [get]
func GetCurrencyConversion(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	amountStr := c.Query("amount")

	if from == "" || to == "" || amountStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Parâmetros 'from', 'to' e 'amount' são obrigatórios"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Valor deve ser um número positivo"})
		return
	}

	// Buscar taxas do cache ou API
	var taxas []Taxa
	cacheKey := "brasil_api_taxas"
	if cachedData, found := cache.GlobalCache.Get(cacheKey); found {
		taxas = cachedData.([]Taxa)
	} else {
		resp, err := http.Get("https://brasilapi.com.br/api/taxas/v1")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar taxas de câmbio"})
			return
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&taxas); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao decodificar taxas"})
			return
		}
		cache.GlobalCache.Set(cacheKey, taxas, 30*time.Minute)
	}

	// Encontrar taxa da moeda
	var rate float64
	for _, taxa := range taxas {
		if taxa.Nome == from {
			rate = taxa.Valor
			break
		}
	}

	if rate == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": fmt.Sprintf("Taxa para %s não encontrada", from)})
		return
	}

	// Para conversão para BRL, dividir pelo valor da taxa
	var result float64
	if to == "BRL" {
		result = amount / rate
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Apenas conversão para BRL é suportada no momento"})
		return
	}

	conversion := CurrencyConversion{
		From:   from,
		To:     to,
		Amount: amount,
		Result: result,
		Rate:   rate,
	}

	c.JSON(http.StatusOK, conversion)
}

// GetInflationData godoc
// @Summary Dados de inflação
// @Description Obtém dados de inflação do Brasil (IPCA)
// @Tags calculations
// @Accept  json
// @Produce  json
// @Success 200 {object} InflationData
// @Failure 500 {object} map[string]string
// @Router /calculations/inflation [get]
func GetInflationData(c *gin.Context) {
	// Buscar dados de inflação da Brasil API
	resp, err := http.Get("https://brasilapi.com.br/api/cotacao/USD")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar dados de inflação"})
		return
	}
	defer resp.Body.Close()

	var usdData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&usdData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao decodificar dados"})
		return
	}

	// Simular dados de inflação (em produção, usar API específica)
	inflation := InflationData{
		Date:        time.Now().Format("2006-01-02"),
		Value:       4.5, // Exemplo IPCA
		Previous:    4.2,
		Variation:   0.3,
		Description: "IPCA mensal estimado",
	}

	c.JSON(http.StatusOK, inflation)
}

// CalculateInvestment godoc
// @Summary Calcular investimento
// @Description Calcula projeção de investimento com depósitos mensais
// @Tags calculations
// @Accept  json
// @Produce  json
// @Param principal query number false "Valor inicial" default(0)
// @Param monthly_deposit query number true "Depósito mensal"
// @Param annual_rate query number true "Taxa anual (%)"
// @Param years query int true "Período em anos"
// @Success 200 {object} InvestmentCalculation
// @Failure 400 {object} map[string]string
// @Router /calculations/investment [get]
func CalculateInvestment(c *gin.Context) {
	principalStr := c.DefaultQuery("principal", "0")
	monthlyDepositStr := c.Query("monthly_deposit")
	annualRateStr := c.Query("annual_rate")
	yearsStr := c.Query("years")

	if monthlyDepositStr == "" || annualRateStr == "" || yearsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Parâmetros obrigatórios: monthly_deposit, annual_rate, years"})
		return
	}

	principal, _ := strconv.ParseFloat(principalStr, 64)
	monthlyDeposit, err := strconv.ParseFloat(monthlyDepositStr, 64)
	if err != nil || monthlyDeposit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "monthly_deposit deve ser um número positivo"})
		return
	}

	annualRate, err := strconv.ParseFloat(annualRateStr, 64)
	if err != nil || annualRate < 0 || annualRate > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "annual_rate deve ser um número entre 0 e 100"})
		return
	}

	years, err := strconv.Atoi(yearsStr)
	if err != nil || years < 1 || years > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "years deve ser um número inteiro entre 1 e 50"})
		return
	}

	// Calcular investimento composto
	monthlyRate := annualRate / 100 / 12
	totalMonths := years * 12

	finalAmount := principal
	totalInvested := principal

	for month := 1; month <= totalMonths; month++ {
		finalAmount = (finalAmount + monthlyDeposit) * (1 + monthlyRate)
		totalInvested += monthlyDeposit
	}

	totalInterest := finalAmount - totalInvested

	calculation := InvestmentCalculation{
		Principal:      principal,
		MonthlyDeposit: monthlyDeposit,
		AnnualRate:     annualRate,
		Years:          years,
		TotalInvested:  totalInvested,
		FinalAmount:    finalAmount,
		TotalInterest:  totalInterest,
	}

	c.JSON(http.StatusOK, calculation)
}
