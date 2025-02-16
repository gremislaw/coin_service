package handler

import (
	"avito_coin/api"
	"avito_coin/internal/service"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// CoinHandler - структура для обработчиков HTTP-запросов
type CoinHandler struct {
	service *service.CoinService
	logger  *logrus.Logger
}

// NewCoinHandler - функция для создания нового обработчика
func NewCoinHandler(e *echo.Echo, service *service.CoinService) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{}) // Используем JSON-формат для логов

	handler := &CoinHandler{
		service: service,
		logger:  logger,
	}

	// Защищенные эндпоинты
	protected := e.Group("") // Группируем защищенные маршруты
	protected.Use(verifyAuth)
	api.RegisterHandlers(protected, handler)
	e.POST("/api/auth", handler.PostApiAuth)
	e.GET("/api/merch/:merch_id", handler.GetMerchPrice)
}

// PostApiAuth - обработчик для авторизации/регистрации пользователя
func (h *CoinHandler) PostApiAuth(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/auth",
		"method":   "POST",
	}).Info("PostApiAuth request received")

	var request api.AuthRequest

	// Парсим тело запроса
	if err := c.Bind(&request); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to parse request body")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Проверяем, есть ли пользователь с таким именем
	if userRow, err := h.service.UserExists(c.Request().Context(), request.Username); err == nil {
		// Пользователь найден, проверяем пароль
		if request.Password != userRow.Password { // В реальной жизни нужно хешировать пароли!
			h.logger.WithFields(logrus.Fields{
				"username": request.Username,
			}).Error("Invalid password")
			errorMessage := "invalid password"
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
				Errors: &errorMessage,
			})
		}

		// Генерация JWT для существующего пользователя
		var resp api.AuthResponse
		token, err := generateJWT(userRow.ID)
		resp.Token = &token
		if err != nil {
			h.logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Failed to generate JWT")
			errorMessage := err.Error()
			return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
				Errors: &errorMessage,
			})
		}

		// Возвращаем успешный ответ с токеном
		h.logger.WithFields(logrus.Fields{
			"user_id":  userRow.ID,
			"username": request.Username,
		}).Info("User authenticated successfully")
		return c.JSON(http.StatusOK, resp)
	}

	// Если пользователь не найден, создаем нового
	h.logger.WithFields(logrus.Fields{
		"username": request.Username,
	}).Info("User not found, creating a new one")

	newID, err := h.service.CreateUser(c.Request().Context(), request.Username, request.Password)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"username": request.Username,
		}).Error("Failed to create user")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Генерация JWT для нового пользователя
	var resp api.AuthResponse
	token, err := generateJWT(newID)
	resp.Token = &token
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to generate JWT")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Возвращаем успешный ответ с токеном
	h.logger.WithFields(logrus.Fields{
		"jwt_user_id": newID,
	}).Info("User created and authenticated successfully")
	return c.JSON(http.StatusOK, resp)
}

// GetApiBuyItem - обработчик для покупки мерча
func (h *CoinHandler) GetApiBuyItem(c echo.Context, item string) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/buy/:item",
		"method":   "GET",
	}).Info("GetApiBuyItem request received")

	merchID, err := strconv.Atoi(item)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"merch_id": c.Param("merch_id"),
		}).Error("Invalid merch ID")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Извлекаем user_id из JWT
	userID := c.Get("jwt_user_id").(int32)

	// Вызываем сервисный слой
	if err := h.service.BuyMerch(c.Request().Context(), userID, int32(merchID)); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"user_id":  userID,
			"merch_id": merchID,
		}).Error("Failed to buy merch")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"merch_id": merchID,
	}).Info("Merch purchased successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "merch purchased successfully"})
}

// PostApiSendCoin - обработчик для перевода монет
func (h *CoinHandler) PostApiSendCoin(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/sendCoin",
		"method":   "POST",
	}).Info("PostApiSendCoin request received")

	var request api.SendCoinRequest

	// Парсим тело запроса
	if err := c.Bind(&request); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to parse request body")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Извлекаем user_id из JWT
	userID := c.Get("jwt_user_id").(int32)

	// Вызываем сервисный слой
	if err := h.service.TransferCoins(c.Request().Context(), userID, request.ToUser, int32(request.Amount)); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":     err.Error(),
			"from_user": userID,
			"to_user":   request.ToUser,
			"amount":    request.Amount,
		}).Error("Failed to transfer coins")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	h.logger.WithFields(logrus.Fields{
		"from_user": userID,
		"to_user":   request.ToUser,
		"amount":    request.Amount,
	}).Info("Coins transferred successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "coins transferred successfully"})
}

// GetMerchPrice - обработчик для получения цены товара
func (h *CoinHandler) GetMerchPrice(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/merch/:merch_id/price",
		"method":   "GET",
	}).Info("GetMerchPrice request received")

	// Извлекаем ID товара из параметров URL
	merchID, err := strconv.Atoi(c.Param("merch_id"))
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"merch_id": c.Param("merch_id"),
		}).Error("Invalid merch ID")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Вызываем сервисный слой для получения цены товара
	price, err := h.service.GetMerchPrice(c.Request().Context(), int32(merchID))
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"merch_id": merchID,
		}).Error("Failed to get merch price")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Логируем успешный ответ
	h.logger.WithFields(logrus.Fields{
		"merch_id": merchID,
		"price":    price,
	}).Info("Merch price retrieved successfully")
	return c.JSON(http.StatusOK, map[string]int32{"price": price})
}

// GetApiInfo - обработчик для получения баланса, покупок и транзакций пользователя
func (h *CoinHandler) GetApiInfo(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/info",
		"method":   "GET",
	}).Info("GetApiInfo request received")

	// Извлекаем user_id из JWT
	userID := c.Get("jwt_user_id").(int32)

	// Получаем баланс пользователя
	balance, err := h.service.GetUserBalance(c.Request().Context(), userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
		}).Error("Failed to get user balance")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Получаем покупки пользователя
	purchases, err := h.service.GetUserPurchases(c.Request().Context(), userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
		}).Error("Failed to get user purchases")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Получаем транзакции пользователя
	transactions, err := h.service.GetTransactions(c.Request().Context(), userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
		}).Error("Failed to get user transactions")
		errorMessage := err.Error()
		return c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Errors: &errorMessage,
		})
	}

	// Логируем успешное выполнение запроса
	h.logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"balance":      balance,
		"purchases":    len(*purchases.Inventory),
		"transactions": len(*transactions.CoinHistory.Received) + len(*transactions.CoinHistory.Sent),
	}).Info("User info retrieved successfully")

	// Возвращаем все данные в одном ответе
	return c.JSON(http.StatusOK, api.InfoResponse{
		Coins:       balance.Coins,
		Inventory:   purchases.Inventory,
		CoinHistory: transactions.CoinHistory,
	})
}
