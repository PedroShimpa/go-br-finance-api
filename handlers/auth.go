package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"go-br-finance-api/config"
	"go-br-finance-api/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key-change-this-in-production")

// validateEmail valida o formato do email
func validateEmail(email string) error {
	if len(email) < 5 || len(email) > 100 {
		return &gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Email deve ter entre 5 e 100 caracteres"}
	}
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email); !matched {
		return &gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Email deve ter um formato válido"}
	}
	return nil
}

// validatePassword valida a força da senha
func validatePassword(password string) error {
	if len(password) < 6 {
		return &gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "Senha deve ter pelo menos 6 caracteres"}
	}
	return nil
}

// Register cria um novo usuário
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	// Validar email
	if err := validateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.(*gin.Error).Meta})
		return
	}

	// Validar senha
	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.(*gin.Error).Meta})
		return
	}

	// Verificar se email já existe
	var existingUser models.User
	err := config.DB.Get(&existingUser, "SELECT id FROM users WHERE email = $1", strings.ToLower(req.Email))
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"erro": "Email já está em uso"})
		return
	}

	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno ao processar senha"})
		return
	}

	// Inserir usuário
	query := "INSERT INTO users (email, password, is_admin) VALUES ($1, $2, $3) RETURNING id"
	var userID int
	err = config.DB.QueryRow(query, strings.ToLower(req.Email), string(hashedPassword), req.IsAdmin).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno ao criar usuário"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"mensagem": "Usuário criado com sucesso", "id": userID})
}

// Login autentica um usuário e retorna JWT
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados inválidos", "detalhes": err.Error()})
		return
	}

	// Validar campos obrigatórios
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email e senha são obrigatórios"})
		return
	}

	// Buscar usuário
	var user models.User
	err := config.DB.Get(&user, "SELECT * FROM users WHERE email = $1", strings.ToLower(req.Email))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	// Verificar senha
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Credenciais inválidas"})
		return
	}

	// Gerar JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"email":    user.Email,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro interno ao gerar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  gin.H{"id": user.ID, "email": user.Email, "is_admin": user.IsAdmin},
	})
}

// AuthMiddleware verifica JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token de autenticação não fornecido"})
			c.Abort()
			return
		}

		// Remover "Bearer " se presente
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Verificar se token não está vazio após remoção
		if strings.TrimSpace(tokenString) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token de autenticação inválido"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token de autenticação inválido"})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "Token de autenticação expirado ou inválido"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["user_id"])
			c.Set("is_admin", claims["is_admin"])
		}

		c.Next()
	}
}

// AdminMiddleware verifica se usuário é admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"erro": "Acesso negado - autenticação requerida"})
			c.Abort()
			return
		}

		if !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"erro": "Acesso negado - apenas administradores"})
			c.Abort()
			return
		}
		c.Next()
	}
}
