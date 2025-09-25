package handlers

import (
	"encoding/json"
	"net/http"

	"go-br-finance-api/config"
	"go-br-finance-api/models"

	"github.com/gin-gonic/gin"
)

type Taxa struct {
	Nome  string  `json:"nome"`
	Valor float64 `json:"valor"`
}

type Resposta struct {
	Taxas         []Taxa                `json:"taxas"`
	Recomendacoes []models.Recomendacao `json:"recomendacoes"`
}

func GetInformacoesFinanceiras(c *gin.Context) {
	// 1. Buscar taxas da Brasil API
	resp, err := http.Get("https://brasilapi.com.br/api/taxas/v1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Não foi possível buscar taxas"})
		return
	}
	defer resp.Body.Close()

	var taxas []Taxa
	if err := json.NewDecoder(resp.Body).Decode(&taxas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao decodificar taxas"})
		return
	}

	// 2. Buscar recomendações no Postgres
	var recomendacoes []models.Recomendacao
	err = config.DB.Select(&recomendacoes, "SELECT * FROM recomendacoes_financeiras")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar recomendações"})
		return
	}

	// 3. Resposta final
	c.JSON(http.StatusOK, Resposta{
		Taxas:         taxas,
		Recomendacoes: recomendacoes,
	})
}
