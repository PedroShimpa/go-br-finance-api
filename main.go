package main

import (
	"log"
	"os"

	"go-br-finance-api/config"
	"go-br-finance-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func runMigrations() {
	sqlBytes, err := os.ReadFile("db/init.sql")
	if err != nil {
		log.Fatal("❌ Erro ao ler arquivo de migração:", err)
	}

	sql := string(sqlBytes)
	_, err = config.DB.Exec(sql)
	if err != nil {
		log.Fatal("❌ Erro ao executar migrações:", err)
	}

	log.Println("✅ Migrações executadas com sucesso")
}

func main() {
	// Carregar variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Conectar banco (commented for testing chat without DB)
	// config.ConnectDB()

	// Conectar Redis
	config.ConnectRedis()

	// Executar migrações (commented for testing)
	// runMigrations()

	// Criar router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.Default())

	// Public endpoints (no authentication required)
	r.GET("/informacoes-financeiras", handlers.GetInformacoesFinanceiras)
	r.GET("/recomendacoes", handlers.GetRecomendacoes)

	// Public endpoints for recommendations
	r.POST("/recomendacoes", handlers.CreateRecomendacao)
	r.PUT("/recomendacoes/:id", handlers.UpdateRecomendacao)
	r.DELETE("/recomendacoes/:id", handlers.DeleteRecomendacao)

	// Calculations endpoints
	r.GET("/calculations/currency", handlers.GetCurrencyConversion)
	r.GET("/calculations/inflation", handlers.GetInflationData)
	r.GET("/calculations/investment", handlers.CalculateInvestment)

	// Chat endpoint
	r.POST("/chat", handlers.ChatWithOllama)

	// Swagger docs endpoint
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Rodar servidor usando porta do env
	port := os.Getenv("PORT")
	if port == "" {
		port = "9091"
	}

	r.Run(":" + port)
}
