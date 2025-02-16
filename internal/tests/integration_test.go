package service_test

import (
	"avito_coin/api"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

// Тест для сценария перевода монет с авторизацией через JWT
func TestSendCoinWithJWT(t *testing.T) {
	// Авторизация для получения токена
	receiver := "test2"
	authRespReceiver, err := authenticateUser(receiver, "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токены
	tokenReceiver := authRespReceiver.Token

	// Проверим баланс до перевода
	initialReceiverBalance := getUserBalance(t, *tokenReceiver)

	// Авторизация для получения токена
	authRespSender, err := authenticateUser("testbro", "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токены
	tokenSender := authRespSender.Token

	// Проверим баланс до перевода
	initialSenderBalance := getUserBalance(t, *tokenSender)

	// Сумма перевода
	amount := int32(2) // Количество монет для перевода

	// Перевод монет
	reqBody := api.SendCoinRequest{
		ToUser: receiver,
		Amount: int(amount),
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Создаем запрос с JWT токеном в заголовке
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/sendCoin", baseURL), bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+*tokenSender) // Добавляем JWT токен в заголовок
	req.Header.Set("Content-Type", "application/json")

	client1 := &http.Client{}
	resp, err := client1.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверим баланс отправителя после перевода
	newSenderBalance := getUserBalance(t, *tokenSender)

	// Убедимся, что баланс отправителя уменьшился на сумму перевода
	assert.Equal(t, initialSenderBalance-amount, newSenderBalance)

	// Cменим сессию
	req.Header.Set("Authorization", "Bearer "+*tokenReceiver) // Добавляем JWT токен в заголовок

	// Проверим баланс отправителя после перевода
	newReceiverBalance := getUserBalance(t, *tokenReceiver)

	// Убедимся, что баланс отправителя уменьшился на сумму перевода
	assert.Equal(t, initialReceiverBalance+amount, newReceiverBalance)
}

// Тест для сценария покупки мерча с авторизацией через JWT
func TestBuyMerchWithJWT(t *testing.T) {
	// Авторизация для получения токена
	authResp, err := authenticateUser("testbro", "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токен
	token := authResp.Token

	// Получаем ID мерча, который мы будем покупать
	merchID := 1

	// Проверим баланс пользователя до покупки
	initialBalance := getUserBalance(t, *token)

	// Создаем запрос с JWT токеном в заголовке
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%v", baseURL, merchID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+*token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем баланс после покупки
	newBalance := getUserBalance(t, *token)

	// Убедитесь, что баланс уменьшился на цену мерча
	merchPrice := getMerchPrice(t, int32(merchID))
	assert.Equal(t, initialBalance-merchPrice, newBalance)
}

// Тест для сценария регистрации и авторизации пользователя
func TestAuthAndRegistration(t *testing.T) {
	// Данные для регистрации нового пользователя
	newUser := api.AuthRequest{
		Username: "newuser",
		Password: "newpassword",
	}

	// Регистрация нового пользователя
	token, err := authenticateUser(newUser.Username, newUser.Password)
	if err != nil {
		t.Fatalf("Failed to register new user: %v", err)
	}
	assert.NotEmpty(t, token, "Token should not be empty after registration")

	// Авторизация существующего пользователя
	existingUser := api.AuthRequest{
		Username: "testbro",
		Password: "test",
	}

	token, err = authenticateUser(existingUser.Username, existingUser.Password)
	if err != nil {
		t.Fatalf("Failed to authenticate existing user: %v", err)
	}
	assert.NotEmpty(t, token, "Token should not be empty after authentication")

	// Попытка авторизации с неверными данными
	invalidUser := api.AuthRequest{
		Username: "testbro",
		Password: "wrongpassword",
	}

	_, err = authenticateUser(invalidUser.Username, invalidUser.Password)
	assert.Error(t, err, "Expected error for invalid credentials")
}

func TestGetBalanceAndTransactionHistory(t *testing.T) {

	// Авторизация пользователя
	authResp, err := authenticateUser("testbro", "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}
	token := authResp.Token

	// Получение информации о балансе и истории транзакций
	infoResponse, err := getInfo(*token)
	if err != nil {
		t.Fatalf("Failed to get user info: %v", err)
	}

	// Проверяем, что баланс не отрицательный
	assert.True(t, *infoResponse.Coins >= 0, "Balance should not be negative")

	// Проверяем, что история транзакций не nil
	assert.NotNil(t, *infoResponse.CoinHistory, "Coin history should not be nil")

	// Проверяем, что инвентарь не nil
	assert.NotNil(t, *infoResponse.Inventory, "Received transactions should not be nil")
}

// getMerchPrice получает цену мерча
func getMerchPrice(t *testing.T, merchID int32) int32 {
	resp, err := http.Get(fmt.Sprintf("%s/api/merch/%d", baseURL, merchID))
	if err != nil {
		t.Fatalf("Failed to get merch price: %v", err)
	}
	defer resp.Body.Close()

	var priceResponse struct {
		Price int32 `json:"price"`
	}
	err = json.NewDecoder(resp.Body).Decode(&priceResponse)
	if err != nil {
		t.Fatalf("Failed to parse merch price: %v", err)
	}

	return priceResponse.Price
}

// getUserBalance получает баланс пользователя
func getUserBalance(t *testing.T, token string) int32 {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/info", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get user balance: %v", err)
	}
	defer resp.Body.Close()

	var infoResponse api.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&infoResponse)
	if err != nil {
		t.Fatalf("Failed to parse info response: %v", err)
	}

	return int32(*infoResponse.Coins)
}

// authenticateUser отправляет запрос для авторизации и получения JWT токена
func authenticateUser(username, password string) (*api.AuthResponse, error) {
	authReq := api.AuthRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %v", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/auth", baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send auth request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var authResp api.AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %v", err)
	}

	return &authResp, nil
}

// getInfo отправляет запрос для получения информации о балансе и истории транзакций
func getInfo(token string) (*api.InfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var infoResponse api.InfoResponse
	err = json.NewDecoder(resp.Body).Decode(&infoResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode info response: %v", err)
	}

	return &infoResponse, nil
}
