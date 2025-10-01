package main

import (
	"log"
	"os"

	"go-br-finance-api/config"
	"go-br-finance-api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Conectar banco
	config.ConnectDB()

	// Criar router
	r := gin.Default()

	// Endpoint
	r.GET("/informacoes-financeiras", handlers.GetInformacoesFinanceiras)

	// Rodar servidor usando porta do env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
