package main

import (
	"go-br-finance-api/config"
	"go-br-finance-api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Conectar banco
	config.ConnectDB()

	// Criar router
	r := gin.Default()

	// Endpoint
	r.GET("/informacoes-financeiras", handlers.GetInformacoesFinanceiras)

	// Rodar servidor
	r.Run(":8080")
}
