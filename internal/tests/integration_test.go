package service_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080"

// Структура для авторизации
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Структура для ответа на авторизацию
type AuthResponse struct {
	Token string `json:"token"`
}

// Структура для запроса перевода монет
type TransferCoinsRequest struct {
	ToUser int32 `json:"to_user"`
	Amount int32 `json:"amount"`
}

// Тест для сценария перевода монет с авторизацией через JWT
func TestSendCoinWithJWT(t *testing.T) {
	// Авторизация для получения токена
	authRespReceiver, err := authenticateUser("test1", "test")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токены
	tokenReceiver := authRespReceiver.Token

	receiverID := int32(1) // ID получателя

	// Проверим баланс до перевода
	initialReceiverBalance := getUserBalance(t)

	// Авторизация для получения токена
	authRespSender, err := authenticateUser("testuser", "testpassword")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токены
	tokenSender := authRespSender.Token

	// Проверим баланс до перевода
	initialSenderBalance := getUserBalance(t)

	// Сумма перевода
	amount := int32(50) // Количество монет для перевода

	// Перевод монет
	reqBody := TransferCoinsRequest{
		ToUser: receiverID,
		Amount: amount,
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
	req.Header.Set("Authorization", "Bearer "+tokenSender) // Добавляем JWT токен в заголовок

	client1 := &http.Client{}
	resp, err := client1.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверим баланс отправителя после перевода
	newSenderBalance := getUserBalance(t)

	// Убедимся, что баланс отправителя уменьшился на сумму перевода
	assert.Equal(t, initialSenderBalance-amount, newSenderBalance)

	// Cменим сессию
	req.Header.Set("Authorization", "Bearer "+tokenReceiver) // Добавляем JWT токен в заголовок

	// Проверим баланс отправителя после перевода
	newReceiverBalance := getUserBalance(t)

	// Убедимся, что баланс отправителя уменьшился на сумму перевода
	assert.Equal(t, initialReceiverBalance-amount, newReceiverBalance)
}

// Структура для пользователя
type User struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
}

// Структура для покупки мерча
type BuyMerchRequest struct {
	UserID  int32 `json:"user_id"`
	MerchID int32 `json:"merch_id"`
}

// Тест для сценария покупки мерча с авторизацией через JWT
func TestBuyMerchWithJWT(t *testing.T) {
	// Авторизация для получения токена
	authResp, err := authenticateUser("testuser", "testpassword")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	// Извлекаем токен
	token := authResp.Token

	// Получаем ID мерча, который мы будем покупать
	merchID := 1

	// Проверим баланс пользователя до покупки
	initialBalance := getUserBalance(t)

	// Создаем запрос с JWT токеном в заголовке
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/buy/%v", baseURL, merchID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус код
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем баланс после покупки
	newBalance := getUserBalance(t)

	// Убедитесь, что баланс уменьшился на цену мерча
	merchPrice := getMerchPrice(t, int32(merchID))
	assert.Equal(t, initialBalance-merchPrice, newBalance)
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
func getUserBalance(t *testing.T) int32 {
	resp, err := http.Get(fmt.Sprintf("%s/api/info", baseURL))
	if err != nil {
		t.Fatalf("Failed to get user balance: %v", err)
	}
	defer resp.Body.Close()

	var infoResponse struct {
		Balance int32 `json:"balance"`
	}
	err = json.NewDecoder(resp.Body).Decode(&infoResponse)
	if err != nil {
		t.Fatalf("Failed to parse balance response: %v", err)
	}

	return infoResponse.Balance
}

// authenticateUser отправляет запрос для авторизации и получения JWT токена
func authenticateUser(username, password string) (*AuthResponse, error) {
	authReq := AuthRequest{
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

	var authResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %v", err)
	}

	return &authResp, nil
}
