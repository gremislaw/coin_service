package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Claims struct {
	UserID   int32  `json:"jwt_user_id"`
	jwt.StandardClaims
}

func generateSecretKey() string {
	// Генерация случайных 32 байтов
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("secret key generated")
	// Преобразуем секретный ключ в base64 для удобства хранения
	return base64.StdEncoding.EncodeToString(secret)
}

var jwtSecret = []byte(generateSecretKey())

func generateJWT(user_id int32) (string, error) {
	claims := &Claims{
		UserID:   user_id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Время жизни токена 24 часа
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Функция для проверки токена
func verifyJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("could not parse claims")
	}
	return claims, nil
}

func verifyAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing token"})
		}

		// Удаляем префикс "Bearer " если он есть
		if len(token) > 6 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Проверяем токен
		claims, err := verifyJWT(token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		// Сохраняем данные о пользователе в контексте
		c.Set("jwt_user_id", claims.UserID)

		return next(c)
	}
}
