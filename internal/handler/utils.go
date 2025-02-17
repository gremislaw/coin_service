package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"avito_coin/api"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func parseAuthRequest(c echo.Context) (*api.AuthRequest, error) {
	var request api.AuthRequest
	if err := c.Bind(&request); err != nil {
		return nil, err
	}

	return &request, nil
}

func parseSendCoinRequest(c echo.Context) (*api.SendCoinRequest, error) {
	var request api.SendCoinRequest
	if err := c.Bind(&request); err != nil {
		logrus.WithFields(logrus.Fields{"error": err.Error()}).Error("Failed to parse request body")
		return nil, err
	}

	return &request, nil
}

func parseMerchID(item string) (int32, error) {
	merchID, err := strconv.ParseInt(item, 10, 32)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err.Error(),
			"merch_id": item,
		}).Error("Invalid merch ID")

		return 0, err
	}

	return int32(merchID), nil
}

func respondWithError(c echo.Context, statusCode int, message string, err error) error {
	logrus.WithFields(logrus.Fields{
		"error": err,
	}).Error(message)

	errorMessage := message

	return c.JSON(statusCode, api.ErrorResponse{
		Errors: &errorMessage,
	})
}

func respondWithToken(c echo.Context, userID int32, logMessage string) error {
	token, err := generateJWT(userID)
	if err != nil {
		return respondWithError(c, http.StatusInternalServerError, "Failed to generate JWT", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info(logMessage)

	return c.JSON(http.StatusOK, api.AuthResponse{
		Token: &token,
	})
}

func respondWithSuccess(c echo.Context, message interface{}, fields logrus.Fields) error {
	logrus.WithFields(fields).Info(message)
	return c.JSON(http.StatusOK, message)
}

func respondWithTransferError(c echo.Context, err error, fromUser int32, toUser string, amount int32) error {
	logrus.WithFields(logrus.Fields{
		"error":     err.Error(),
		"from_user": fromUser,
		"to_user":   toUser,
		"amount":    amount,
	}).Error("Failed to transfer coins")

	return respondWithError(c, http.StatusInternalServerError, "Failed to transfer coins", err)
}

func validatePassword(inputPassword, storedPassword string) bool {
	return inputPassword == storedPassword
}

func validateAmount(value int) (int32, error) {
	if value > math.MaxInt32 || value < math.MinInt32 {
		return 0, fmt.Errorf("amount value %d is out of range for int32", value)
	}

	return int32(value), nil
}

func extractUserID(c echo.Context) (int32, error) {
	userID, ok := c.Get("jwt_user_id").(int32)
	if !ok {
		logrus.Error("Failed to extract user ID from JWT")
		return 0, fmt.Errorf("invalid user ID")
	}

	return userID, nil
}

func extractMerchID(c echo.Context) (int32, error) {
	merchID, err := strconv.ParseInt(c.Param("merch_id"), 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(merchID), nil
}

// getUserInfo получает баланс, покупки и транзакции пользователя.
func (h *CoinHandler) getUserInfo(ctx context.Context, userID int32) (*api.InfoResponse, error) {
	var info api.InfoResponse

	balance, err := h.service.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	purchases, err := h.service.GetUserPurchases(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user purchases: %w", err)
	}

	transactions, err := h.service.GetTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
	}

	info.Coins = balance.Coins
	info.Inventory = purchases.Inventory
	info.CoinHistory = transactions.CoinHistory

	return &info, nil
}
