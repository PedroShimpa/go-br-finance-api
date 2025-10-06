package main

import (
	"log"
	"os"

	"go-br-finance-api/config"
	"go-br-finance-api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Public endpoints (no authentication required)
	r.GET("/informacoes-financeiras", handlers.GetInformacoesFinanceiras)
	r.GET("/recomendacoes", handlers.GetRecomendacoes)

	// Authentication endpoints
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected endpoints (require authentication and admin role)
	protected := r.Group("/")
	protected.Use(handlers.AuthMiddleware())
	protected.Use(handlers.AdminMiddleware())
	{
		protected.POST("/recomendacoes", handlers.CreateRecomendacao)
		protected.PUT("/recomendacoes/:id", handlers.UpdateRecomendacao)
		protected.DELETE("/recomendacoes/:id", handlers.DeleteRecomendacao)
	}

	// Calculations endpoints
	r.GET("/calculations/currency", handlers.GetCurrencyConversion)
	r.GET("/calculations/inflation", handlers.GetInflationData)
	r.GET("/calculations/investment", handlers.CalculateInvestment)

	// Swagger docs endpoint
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Rodar servidor usando porta do env
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	r.Run(":" + port)
}
