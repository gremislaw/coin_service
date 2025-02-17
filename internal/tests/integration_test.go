package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"avito_coin/api"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

// Тест для сценария перевода монет с авторизацией через JWT.
func TestSendCoinWithJWT(t *testing.T) {
	// Авторизация для получения токена
	receiver := "test2"

	authRespReceiver, err := authenticateUser(t, receiver, "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токены
	tokenReceiver := authRespReceiver.Token

	// Проверим баланс до перевода
	initialReceiverBalance := getUserBalance(t, *tokenReceiver)

	// Авторизация для получения токена
	authRespSender, err := authenticateUser(t, "testbro", "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токены
	tokenSender := authRespSender.Token

	// Проверим баланс до перевода
	initialSenderBalance := getUserBalance(t, *tokenSender)

	// Сумма перевода
	amount := 2 // Количество монет для перевода

	// Перевод монет
	reqBody := api.SendCoinRequest{
		ToUser: receiver,
		Amount: amount,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Создаем запрос с JWT токеном в заголовке
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("%s/api/sendCoin", baseURL),
		bytes.NewBuffer(body),
	)
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

// Тест для сценария покупки мерча с авторизацией через JWT.
func TestBuyMerchWithJWT(t *testing.T) {
	// Авторизация для получения токена
	authResp, err := authenticateUser(t, "testbro", "test")
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
	req, err := http.NewRequestWithContext(
		context.Background(), // Или переданный контекст
		http.MethodGet,
		fmt.Sprintf("%s/api/buy/%v", baseURL, merchID),
		nil,
	)
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
	merchPrice := getMerchPrice(t, merchID)
	assert.Equal(t, initialBalance-merchPrice, newBalance)
}

// Тест для сценария регистрации и авторизации пользователя.
func TestAuthAndRegistration(t *testing.T) {
	// Данные для регистрации нового пользователя
	newUser := api.AuthRequest{
		Username: "newuser",
		Password: "newpassword",
	}

	// Регистрация нового пользователя
	token, err := authenticateUser(t, newUser.Username, newUser.Password)
	if err != nil {
		t.Fatalf("Failed to register new user: %v", err)
	}

	assert.NotEmpty(t, token, "Token should not be empty after registration")

	// Авторизация существующего пользователя
	existingUser := api.AuthRequest{
		Username: "testbro",
		Password: "test",
	}

	token, err = authenticateUser(t, existingUser.Username, existingUser.Password)
	if err != nil {
		t.Fatalf("Failed to authenticate existing user: %v", err)
	}

	assert.NotEmpty(t, token, "Token should not be empty after authentication")

	// Попытка авторизации с неверными данными
	invalidUser := api.AuthRequest{
		Username: "testbro",
		Password: "wrongpassword",
	}

	_, err = authenticateUser(t, invalidUser.Username, invalidUser.Password)
	assert.Error(t, err, "Expected error for invalid credentials")
}

func TestGetBalanceAndTransactionHistory(t *testing.T) {
	// Авторизация пользователя
	authResp, err := authenticateUser(t, "testbro", "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	token := authResp.Token

	// Получение информации о балансе и истории транзакций
	infoResponse, err := getInfo(t, *token)
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

// getMerchPrice получает цену мерча.
func getMerchPrice(t *testing.T, merchID int) int {
	t.Helper() // Сообщаем Go, что это вспомогательная функция

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("%s/api/merch/%d", baseURL, merchID),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
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

	return int(priceResponse.Price)
}

// getUserBalance получает баланс пользователя.
func getUserBalance(t *testing.T, token string) int {
	t.Helper() // Сообщаем Go, что это вспомогательная функция

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		baseURL+"/api/info",
		nil,
	)
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

	return *infoResponse.Coins
}

// authenticateUser отправляет запрос для авторизации и получения JWT токена.
func authenticateUser(t *testing.T, username, password string) (*api.AuthResponse, error) {
	t.Helper() // Сообщаем Go, что это вспомогательная функция

	authReq := api.AuthRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		fmt.Sprintf("%s/api/auth", baseURL),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send auth request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var authResp api.AuthResponse

	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	return &authResp, nil
}

// getInfo отправляет запрос для получения информации о балансе и истории транзакций.
func getInfo(t *testing.T, token string) (*api.InfoResponse, error) {
	t.Helper() // Сообщаем Go, что это вспомогательная функция

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/api/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var infoResponse api.InfoResponse

	err = json.NewDecoder(resp.Body).Decode(&infoResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode info response: %w", err)
	}

	return &infoResponse, nil
}
