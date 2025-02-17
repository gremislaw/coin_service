package handler

import (
	"net/http"
	"time"

	"avito_coin/api"
	"avito_coin/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// CoinHandler - структура для обработчиков HTTP-запросов.
type CoinHandler struct {
	service *service.CoinService
	logger  *logrus.Logger
}

// NewCoinHandler - функция для создания нового обработчика.
func NewCoinHandler(e *echo.Echo, service *service.CoinService) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{}) // Используем JSON-формат для логов

	handler := &CoinHandler{
		service: service,
		logger:  logger,
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			time.Sleep(50 * time.Millisecond) // Задержка в 50 мс
			return next(c)
		}
	})

	// Защищенные эндпоинты
	// Общая группа API (без middleware)
	public := e.Group("")

	protected := public.Group("") // Группируем защищенные маршруты
	protected.Use(verifyAuth)

	api.RegisterHandlers(public, protected, handler)
	public.GET("/api/merch/:merch_id", handler.GetMerchPrice) // своя ручка (посчитал нужным)
}

// PostApiAuth - обработчик для авторизации/регистрации пользователя.
func (h *CoinHandler) PostAPIAuth(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/auth",
		"method":   "POST",
	}).Info("PostApiAuth request received")

	// Парсим запрос
	request, err := parseAuthRequest(c)
	if err != nil {
		return respondWithError(c, http.StatusBadRequest, "Failed to parse request body", err)
	}

	// Проверяем, существует ли пользователь
	user, err := h.service.UserExists(c.Request().Context(), request.Username)
	if err == nil {
		// Проверяем пароль
		if !validatePassword(request.Password, user.Password) {
			return respondWithError(c, http.StatusUnauthorized, "Invalid password", nil)
		}

		// Генерируем JWT и отправляем ответ
		return respondWithToken(c, user.ID, "User authenticated successfully")
	}

	// Регистрируем нового пользователя
	newUserID, err := h.service.CreateUser(c.Request().Context(), request.Username, request.Password)
	if err != nil {
		return respondWithError(c, http.StatusInternalServerError, "Failed to create user", err)
	}

	// Генерируем JWT и отправляем ответ
	return respondWithToken(c, newUserID, "User created and authenticated successfully")
}

// GetApiBuyItem - обработчик для покупки мерча.
func (h *CoinHandler) GetAPIBuyItem(c echo.Context, item string) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/buy/:item",
		"method":   "GET",
	}).Info("GetApiBuyItem request received")

	// Получаем merchID
	merchID, err := parseMerchID(item)
	if err != nil {
		return respondWithError(c, http.StatusBadRequest, "Invalid merch ID", err)
	}

	// Получаем userID из JWT
	userID, err := extractUserID(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, "Invalid user ID", err)
	}

	// Вызываем сервисный слой
	if err := h.service.BuyMerch(c.Request().Context(), userID, merchID); err != nil {
		return respondWithError(c, http.StatusInternalServerError, "Failed to buy merch", err)
	}

	// Логируем и возвращаем успешный ответ
	h.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"merch_id": merchID,
	}).Info("Merch purchased successfully")

	return respondWithSuccess(c, "Merch purchased successfully", logrus.Fields{
		"user_id":  userID,
		"merch_id": merchID,
	})
}

// PostApiSendCoin - обработчик для перевода монет.
func (h *CoinHandler) PostAPISendCoin(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/sendCoin",
		"method":   "POST",
	}).Info("PostApiSendCoin request received")

	// Парсим тело запроса
	request, err := parseSendCoinRequest(c)
	if err != nil {
		return respondWithError(c, http.StatusBadRequest, "Invalid request body", err)
	}

	// Проверяем диапазон значения amount
	amount, err := validateAmount(request.Amount)
	if err != nil {
		return respondWithError(c, http.StatusInternalServerError, err.Error(), nil)
	}

	// Извлекаем user_id из JWT
	userID, err := extractUserID(c)
	if err != nil {
		return respondWithError(c, http.StatusUnauthorized, "Invalid user ID", err)
	}

	// Вызываем сервисный слой
	if err := h.service.TransferCoins(c.Request().Context(), userID, request.ToUser, amount); err != nil {
		return respondWithTransferError(c, err, userID, request.ToUser, amount)
	}

	// Логируем и отправляем успешный ответ
	return respondWithSuccess(c, "Coins transferred successfully", logrus.Fields{
		"from_user": userID,
		"to_user":   request.ToUser,
		"amount":    request.Amount,
	})
}

// GetMerchPrice - обработчик для получения цены товара.
func (h *CoinHandler) GetMerchPrice(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/merch/:merch_id/price",
		"method":   "GET",
	}).Info("GetMerchPrice request received")

	// Извлекаем merch_id из параметров
	merchID, err := extractMerchID(c)
	if err != nil {
		return respondWithError(c, http.StatusInternalServerError, "Invalid merch ID", err)
	}

	// Получаем цену товара через сервисный слой
	price, err := h.service.GetMerchPrice(c.Request().Context(), merchID)
	if err != nil {
		return respondWithError(c, http.StatusInternalServerError, "Failed to get merch price", err)
	}

	// Логируем успешный ответ и возвращаем результат
	return respondWithSuccess(c, map[string]int32{"price": price}, logrus.Fields{
		"merch_id": merchID,
		"price":    price,
	})
}

// GetApiInfo - обработчик для получения баланса, покупок и транзакций пользователя.
func (h *CoinHandler) GetAPIInfo(c echo.Context) error {
	h.logger.WithFields(logrus.Fields{
		"endpoint": "/info",
		"method":   "GET",
	}).Info("GetApiInfo request received")

	// Извлекаем user_id из JWT
	userID := c.Get("jwt_user_id").(int32)

	// Получаем информацию о пользователе
	info, err := h.getUserInfo(c.Request().Context(), userID)
	if err != nil {
		return respondWithError(c, http.StatusInternalServerError, "Failed to retrieve user info", err)
	}

	// Логируем успешное выполнение запроса
	return respondWithSuccess(c, info, logrus.Fields{
		"user_id":      userID,
		"balance":      info.Coins,
		"purchases":    len(*info.Inventory),
		"transactions": len(*info.CoinHistory.Received) + len(*info.CoinHistory.Sent),
	})
}
