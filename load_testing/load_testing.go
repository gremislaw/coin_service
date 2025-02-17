package main

import (
	"avito_coin/api"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Генерация безопасной случайной строки фиксированной длины.
func randString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var result []byte

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err)
		}

		result = append(result, chars[n.Int64()])
	}

	return string(result)
}

// Генерация случайного числа в заданном диапазоне (безопасно).
func randInt(minn, maxx int) int {
	if minn >= maxx {
		panic("min > max")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxx-minn+1)))
	if err != nil {
		panic(err)
	}

	return int(n.Int64()) + minn
}

// Авторизация пользователя, получение JWT.
func authenticateUser(userID int) (string, error) {
	username := fmt.Sprintf("test%d", userID)
	password := "test"
	credentials := map[string]string{
		"username": username,
		"password": password,
	}

	body, _ := json.Marshal(credentials)

	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"http://localhost:8080/api/auth",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate, status code: %d", resp.StatusCode)
	}

	// Читаем ответ и извлекаем токен
	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	token, ok := response["token"]
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

// Добавляем JWT в заголовки запроса.
func addJWTToHeader(t *vegeta.Target, userID int) error {
	token, _ := authenticateUser(userID)

	t.Header = map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + token},
	}

	return nil
}

// Функция для выполнения запроса на создание пользователя.
func testCreateUser(t *vegeta.Target) error {
	t.Method = "POST"
	t.URL = "http://localhost:8080/api/auth"
	username := randString(randInt(5, 15))
	password := randString(randInt(8, 15))
	data := map[string]string{
		"username": username,
		"password": password,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	t.Body = body
	t.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	// Перед выполнением запроса добавим авторизацию
	userID := randInt(1, 100000) // Генерация случайного userID для теста
	if err := addJWTToHeader(t, userID); err != nil {
		return err
	}

	return nil
}

// Функция для выполнения запроса на покупку товара.
func testBuyMerchTarget(t *vegeta.Target) error {
	t.Method = "GET"
	userID := randInt(1, 100000)
	merchID := randInt(1, 10)
	t.URL = fmt.Sprintf("http://localhost:8080/api/buy/%v", merchID)

	t.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	// Перед выполнением запроса добавим авторизацию
	if err := addJWTToHeader(t, userID); err != nil {
		return err
	}

	return nil
}

// Функция для выполнения запроса на перевод монет.
func testSendCoinTarget(t *vegeta.Target) error {
	t.Method = "POST"
	t.URL = "http://localhost:8080/api/sendCoin"
	fromUserID := randInt(1, 100000)
	toUser := randInt(1, 100000)
	amount := randInt(10, 100)
	data := api.SendCoinRequest{
		ToUser: "test"+strconv.Itoa(toUser),
		Amount:  amount,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	t.Body = body
	t.Header = map[string][]string{
		"Content-Type": {"application/json"},
	}

	// Перед выполнением запроса добавим авторизацию
	if err := addJWTToHeader(t, fromUserID); err != nil {
		return err
	}

	return nil
}

// Функция для выполнения запроса на получение баланса пользователя.
func testGetAPIInfo(t *vegeta.Target) error {
	userID := randInt(1, 100000)
	t.Method = "GET"
	t.URL = "http://localhost:8080/api/info"

	// Перед выполнением запроса добавим авторизацию
	if err := addJWTToHeader(t, userID); err != nil {
		return err
	}

	return nil
}

func runTest(
	attacker *vegeta.Attacker,
	rate vegeta.Rate,
	duration time.Duration,
) {

	var metrics vegeta.Metrics
	for res := range attacker.Attack(generateTargets(), rate, duration, "mixed testing") {
		metrics.Add(res)
	}

	metrics.Close()

	// Выводим результаты теста
	fmt.Println("Test Results:\n")
	printMetrics(metrics)
}

// Функция для генерации запросов (разных типов)
func generateTargets() vegeta.Targeter {
	targets := []func(t *vegeta.Target) error{
		testBuyMerchTarget,
		testGetAPIInfo,
		testSendCoinTarget,
	}

	return targets[time.Now().UnixNano()%int64(len(targets))]
}

// Функция для вывода метрик теста.
func printMetrics(metrics vegeta.Metrics) {
	fmt.Printf("Requests: %d\n", metrics.Requests)
	fmt.Printf("Success Rate: %.2f%%\n", metrics.Success*100)
	fmt.Printf("Latency (mean): %s\n", metrics.Latencies.Mean)
	fmt.Printf("Latency (95th percentile): %s\n", metrics.Latencies.P95)
	fmt.Printf("Latency (99th percentile): %s\n", metrics.Latencies.P99)
	fmt.Printf("Bytes In (mean): %.2f\n", metrics.BytesIn.Mean)
	fmt.Printf("Bytes Out (mean): %.2f\n", metrics.BytesOut.Mean)
}

func main() {
	// Настройка интенсивности запросов
	pacer := vegeta.ConstantPacer{Freq: 1000, Per: time.Second} // 1000 запросов в секунду


	// Длительность теста
	duration := 10 * time.Second

	// Создаем атакующего (attacker)
	attacker := vegeta.NewAttacker()

	_ = testCreateUser

	runTest(attacker, pacer, duration)
}
