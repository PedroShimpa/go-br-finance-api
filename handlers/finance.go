package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go-br-finance-api/cache"
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

// GetInformacoesFinanceiras godoc
// @Summary Obter informações financeiras
// @Description Obtém informações financeiras incluindo taxas e recomendações
// @Tags finance
// @Accept  json
// @Produce  json
// @Success 200 {object} Resposta
// @Failure 500 {object} map[string]string
// @Router /informacoes-financeiras [get]
func GetInformacoesFinanceiras(c *gin.Context) {
	var taxas []Taxa

	// Tentar obter taxas do cache primeiro
	cacheKey := "brasil_api_taxas"
	if cachedData, found := cache.GlobalCache.Get(cacheKey); found {
		taxas = cachedData.([]Taxa)
	} else {
		// Buscar taxas da Brasil API
		resp, err := http.Get("https://brasilapi.com.br/api/taxas/v1")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Não foi possível buscar taxas"})
			return
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&taxas); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao decodificar taxas"})
			return
		}

		// Cache por 30 minutos
		cache.GlobalCache.Set(cacheKey, taxas, 30*time.Minute)
	}

	// Buscar recomendações no Postgres
	var recomendacoes []models.Recomendacao
	err := config.DB.Select(&recomendacoes, "SELECT * FROM recomendacoes_financeiras")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar recomendações"})
		return
	}

	// Resposta final
	c.JSON(http.StatusOK, Resposta{
		Taxas:         taxas,
		Recomendacoes: recomendacoes,
	})
}

// GetRecomendacoes godoc
// @Summary Listar recomendações
// @Description Obtém todas as recomendações financeiras
// @Tags recommendations
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Recomendacao
// @Failure 500 {object} map[string]string
// @Router /recomendacoes [get]
func GetRecomendacoes(c *gin.Context) {
	var recomendacoes []models.Recomendacao
	err := config.DB.Select(&recomendacoes, "SELECT * FROM recomendacoes_financeiras")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar recomendações"})
		return
	}

	c.JSON(http.StatusOK, recomendacoes)
}

// CreateRecomendacao godoc
// @Summary Criar recomendação
// @Description Cria uma nova recomendação financeira (requer autenticação de admin)
// @Tags recommendations
// @Accept  json
// @Produce  json
// @Param request body models.Recomendacao true "Dados da recomendação"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /recomendacoes [post]
func CreateRecomendacao(c *gin.Context) {
	var recomendacao models.Recomendacao
	if err := c.ShouldBindJSON(&recomendacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	// Validação básica
	if len(recomendacao.Titulo) < 3 || len(recomendacao.Titulo) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Título deve ter entre 3 e 200 caracteres"})
		return
	}

	if len(recomendacao.Descricao) < 10 || len(recomendacao.Descricao) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Descrição deve ter entre 10 e 1000 caracteres"})
		return
	}

	query := "INSERT INTO recomendacoes_financeiras (titulo, descricao) VALUES ($1, $2) RETURNING id"
	var id int
	err := config.DB.QueryRow(query, recomendacao.Titulo, recomendacao.Descricao).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar recomendação"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensagem": "Recomendação criada com sucesso", "id": id})
}

// UpdateRecomendacao godoc
// @Summary Atualizar recomendação
// @Description Atualiza uma recomendação financeira existente (requer autenticação de admin)
// @Tags recommendations
// @Accept  json
// @Produce  json
// @Param id path int true "ID da recomendação"
// @Param request body models.Recomendacao true "Dados atualizados da recomendação"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /recomendacoes/{id} [put]
func UpdateRecomendacao(c *gin.Context) {
	id := c.Param("id")
	var recomendacao models.Recomendacao
	if err := c.ShouldBindJSON(&recomendacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	// Validação básica
	if len(recomendacao.Titulo) < 3 || len(recomendacao.Titulo) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Título deve ter entre 3 e 200 caracteres"})
		return
	}

	if len(recomendacao.Descricao) < 10 || len(recomendacao.Descricao) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Descrição deve ter entre 10 e 1000 caracteres"})
		return
	}

	query := "UPDATE recomendacoes_financeiras SET titulo = $1, descricao = $2 WHERE id = $3"
	result, err := config.DB.Exec(query, recomendacao.Titulo, recomendacao.Descricao, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar recomendação"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Recomendação não encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Recomendação atualizada com sucesso"})
}

// DeleteRecomendacao godoc
// @Summary Deletar recomendação
// @Description Remove uma recomendação financeira (requer autenticação de admin)
// @Tags recommendations
// @Accept  json
// @Produce  json
// @Param id path int true "ID da recomendação"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /recomendacoes/{id} [delete]
func DeleteRecomendacao(c *gin.Context) {
	id := c.Param("id")

	query := "DELETE FROM recomendacoes_financeiras WHERE id = $1"
	result, err := config.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar recomendação"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Recomendação não encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "Recomendação deletada com sucesso"})
}
