package service

import (
	"context"
	"fmt"

	"avito_coin/api"
	"avito_coin/internal/db"
	"avito_coin/internal/repository"
)

// CoinService - сервис для работы с монетками и мерчем
type CoinService struct {
	repo repository.Repository
}

// NewCoinService - функция для создания нового сервиса
func NewCoinService(repo repository.Repository) *CoinService {
	return &CoinService{
		repo: repo,
	}
}

// CreateUser - создание пользователя
func (s *CoinService) CreateUser(ctx context.Context, username, password string) (int32, error) {
	return s.repo.CreateUser(ctx, username, password)
}

// CreateMerch - создание мерча
func (s *CoinService) CreateMerch(ctx context.Context, name string, price int32) error {
	return s.repo.CreateMerch(ctx, name, price)
}

// BuyMerch - покупка мерча пользователем
func (s *CoinService) BuyMerch(ctx context.Context, userID, merchID int32) error {
	// Проверяем, существует ли пользователь и мерч
	balance, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	price, err := s.repo.GetMerchPrice(ctx, merchID)
	if err != nil {
		return fmt.Errorf("merch not found: %w", err)
	}

	// Проверяем, достаточно ли монет для покупки
	if balance < price {
		return fmt.Errorf("insufficient balance for purchase")
	}

	// Проверяем, что баланс не станет отрицательным после покупки
	newBalance := balance - price
	if newBalance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	// Выполняем покупку через репозиторий
	return s.repo.BuyMerch(ctx, userID, merchID)
}

// TransferCoins - перевод монет от одного пользователя к другому
func (s *CoinService) TransferCoins(ctx context.Context, fromUserID int32, toUser string, amount int32) error {
	toUserData, err := s.repo.UserExists(ctx, toUser)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if fromUserID == toUserData.ID {
		return fmt.Errorf("sender and receiver cannot be the same")
	}

	// Проверяем, существуют ли пользователи
	senderBalance, err := s.repo.GetUserBalance(ctx, int32(fromUserID))
	if err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	_, err = s.repo.GetUserBalance(ctx, toUserData.ID)
	if err != nil {
		return fmt.Errorf("receiver not found: %w", err)
	}

	// Проверяем, достаточно ли монет для перевода
	if senderBalance < amount {
		return fmt.Errorf("insufficient balance to transfer")
	}

	// Проверяем, что баланс отправителя не станет отрицательным после перевода
	newSenderBalance := senderBalance - amount
	if newSenderBalance < 0 {
		return fmt.Errorf("sender balance cannot be negative")
	}

	// Выполняем перевод через репозиторий
	return s.repo.TransferCoins(ctx, int32(fromUserID), toUserData.ID, int32(amount))
}

// GetMerchPrice - получение цены мерча
func (s *CoinService) GetMerchPrice(ctx context.Context, id int32) (int32, error) {
	// Проверяем, существует ли пользователь
	balance, err := s.repo.GetMerchPrice(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("merch not found: %w", err)
	}

	return balance, err
}

// GetUserBalance - получение баланса пользователя
func (s *CoinService) GetUserBalance(ctx context.Context, userID int32) (*api.InfoResponse, error) {
	// Проверяем, существует ли пользователь
	balance, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return &api.InfoResponse{}, fmt.Errorf("user not found: %w", err)
	}

	resBalance := int(balance)
	return &api.InfoResponse{Coins: &resBalance}, err
}

// GetUserPurchases - получение всех покупок пользователя
func (s *CoinService) GetUserPurchases(ctx context.Context, userID int32) (*api.InfoResponse, error) {
	// Получаем покупки из репозитория
	purchases, err := s.repo.GetUserPurchases(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user purchases: %w", err)
	}

	// Создаем структуру для ответа
	infoResponse := &api.InfoResponse{
		Inventory: &[]struct {
			Quantity *int    `json:"quantity,omitempty"`
			Type     *string `json:"type,omitempty"`
		}{},
	}

	// Обрабатываем покупки
	itemCounts := make(map[string]int)
	for _, purchase := range purchases {
		// Подсчитываем количество каждого типа предмета
		itemCounts[purchase.Name]++
	}

	// Временная переменная для хранения списка предметов в инвентаре
	var inventoryList []struct {
		Quantity *int    `json:"quantity,omitempty"`
		Type     *string `json:"type,omitempty"`
	}

	// Обрабатываем покупки
	for itemType, quantity := range itemCounts {
		inventoryList = append(inventoryList, struct {
			Quantity *int    `json:"quantity,omitempty"`
			Type     *string `json:"type,omitempty"`
		}{
			Quantity: &quantity,
			Type:     &itemType,
		})
	}

	// Заполняем структуру ответа
	infoResponse.Inventory = &inventoryList

	return infoResponse, nil
}

// GetTransactions - получение списка всех транзакций пользователя, сгруппированных по полученным и отправленным монетам
func (s *CoinService) GetTransactions(ctx context.Context, userID int32) (*api.InfoResponse, error) {
	// Получаем транзакции из репозитория
	transactions, err := s.repo.GetTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Создаем структуру для ответа
	infoResponse := &api.InfoResponse{
		CoinHistory: &struct {
			Received *[]struct {
				Amount   *int    `json:"amount,omitempty"`
				FromUser *string `json:"fromUser,omitempty"`
			} `json:"received,omitempty"`
			Sent *[]struct {
				Amount *int    `json:"amount,omitempty"`
				ToUser *string `json:"toUser,omitempty"`
			} `json:"sent,omitempty"`
		}{},
	}

	// Временные переменные для хранения полученных и отправленных транзакций
	var received []struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	}
	var sent []struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	}

	// Обрабатываем транзакции
	for _, tx := range transactions {
		if tx.ToUser.Int32 == userID {
			// Это полученные монеты
			amount := int(tx.Amount)
			fromUser := tx.FromUser.Int32
			fromUserName := fmt.Sprintf("user%d", fromUser) // Здесь можно получить имя пользователя из БД
			received = append(received, struct {
				Amount   *int    `json:"amount,omitempty"`
				FromUser *string `json:"fromUser,omitempty"`
			}{
				Amount:   &amount,
				FromUser: &fromUserName,
			})
		} else if tx.FromUser.Int32 == userID {
			// Это отправленные монеты
			amount := int(tx.Amount)
			toUser := tx.ToUser.Int32
			toUserName := fmt.Sprintf("user%d", toUser) // Здесь можно получить имя пользователя из БД
			sent = append(sent, struct {
				Amount *int    `json:"amount,omitempty"`
				ToUser *string `json:"toUser,omitempty"`
			}{
				Amount: &amount,
				ToUser: &toUserName,
			})
		}
	}

	// Заполняем структуру ответа
	infoResponse.CoinHistory.Received = &received
	infoResponse.CoinHistory.Sent = &sent

	return infoResponse, nil
}

// UpdateUserBalance - обновление баланса пользователя
func (s *CoinService) UpdateUserBalance(ctx context.Context, userID int32, balance int32) error {
	// Проверяем, существует ли пользователь
	_, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем, что новый баланс не отрицательный
	if balance < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	// Обновляем баланс через репозиторий
	return s.repo.UpdateUserBalance(ctx, userID, balance)
}

// UserExists - проверка на сущестование пользователя
func (s *CoinService) UserExists(ctx context.Context, username string) (db.UserExistsRow, error) {
	return s.repo.UserExists(ctx, username)
}
