package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTConfig struct {
	TokenExpiration time.Duration
	TokenSecret     []byte
}

var Config JWTConfig

func LoadTokenConfig() {
	Config = JWTConfig{
		TokenExpiration: 24 * time.Hour,
		TokenSecret:     []byte(os.Getenv("JWT_SECRET_KEY")),
	}
}

func GenerateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(Config.TokenExpiration).Unix(),
	})

	t, err := token.SignedString(Config.TokenSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}
	return t, nil
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return Config.TokenSecret, nil
	})
}

func IsTokenValid(tokenString string) bool {
	token, err := ValidateToken(tokenString)
	return err == nil && token.Valid
}

func GetUserIDFromToken(tokenString string) (uuid.UUID, error) {
	token, err := ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id not found in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user_id format: %v", err)
	}

	return userID, nil
}

// GetUserIDFromContext извлекает user ID из cookie или Bearer токена в Fiber контексте
func GetUserIDFromContext(c *fiber.Ctx) (uuid.UUID, error) {
	// Сначала пробуем получить токен из cookie
	tokenString := c.Cookies("auth_token")

	// Если cookie нет, пробуем получить из Authorization header
	if tokenString == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
	}

	if tokenString == "" {
		return uuid.Nil, fmt.Errorf("auth token not found in cookie or authorization header")
	}

	return GetUserIDFromToken(tokenString)
}
