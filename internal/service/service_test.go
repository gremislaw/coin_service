package service_test

import (
	"context"
	"database/sql"
	"testing"

	"avito_coin/internal/db"
	"avito_coin/internal/service"

	"github.com/stretchr/testify/assert"
)

// MockRepository - мок-репозиторий для тестирования
type MockRepository struct {
	CreateUserFunc          func(ctx context.Context, username, password string) (int32, error)
	CreateMerchFunc         func(ctx context.Context, name string, price int32) error
	BuyMerchFunc            func(ctx context.Context, userID, merchID int32) error
	GetMerchPriceFunc       func(ctx context.Context, merchID int32) (int32, error)
	TransferCoinsFunc       func(ctx context.Context, fromUser, toUser, amount int32) error
	GetUserBalanceFunc      func(ctx context.Context, userID int32) (int32, error)
	GetUserPurchasesFunc    func(ctx context.Context, userID int32) ([]db.GetUserPurchasesRow, error)
	GetTransactionsFunc     func(ctx context.Context, userID int32) ([]db.GetTransactionsRow, error)
	UpdateUserBalanceFunc   func(ctx context.Context, userID int32, balance int32) error
	UserExistsFunc          func(ctx context.Context, username string) (db.UserExistsRow, error)
}

func (m *MockRepository) CreateUser(ctx context.Context, username, password string) (int32, error) {
	return m.CreateUserFunc(ctx, username, password)
}

func (m *MockRepository) CreateMerch(ctx context.Context, name string, price int32) error {
	return m.CreateMerchFunc(ctx, name, price)
}

func (m *MockRepository) BuyMerch(ctx context.Context, userID, merchID int32) error {
	return m.BuyMerchFunc(ctx, userID, merchID)
}

func (m *MockRepository) GetMerchPrice(ctx context.Context, merchID int32) (int32, error) {
	return m.GetMerchPriceFunc(ctx, merchID)
}

func (m *MockRepository) TransferCoins(ctx context.Context, fromUser, toUser, amount int32) error {
	return m.TransferCoinsFunc(ctx, fromUser, toUser, amount)
}

func (m *MockRepository) GetUserBalance(ctx context.Context, userID int32) (int32, error) {
	return m.GetUserBalanceFunc(ctx, userID)
}

func (m *MockRepository) GetUserPurchases(ctx context.Context, userID int32) ([]db.GetUserPurchasesRow, error) {
	return m.GetUserPurchasesFunc(ctx, userID)
}

func (m *MockRepository) GetTransactions(ctx context.Context, userID int32) ([]db.GetTransactionsRow, error) {
	return m.GetTransactionsFunc(ctx, userID)
}

func (m *MockRepository) UpdateUserBalance(ctx context.Context, userID int32, balance int32) error {
	return m.UpdateUserBalanceFunc(ctx, userID, balance)
}

func (m *MockRepository) UserExists(ctx context.Context, username string) (db.UserExistsRow, error) {
	return m.UserExistsFunc(ctx, username)
}

func TestCreateUser(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockRepository{
		CreateUserFunc: func(ctx context.Context, username, password string) (int32, error) {
			return 1, nil // Возвращаем ID нового пользователя
		},
	}

	// Создаем сервис с мок-репозиторием
	coinService := service.NewCoinService(mockRepo)

	// Вызываем метод CreateUser
	userID, err := coinService.CreateUser(context.Background(), "testuser", "testpassword")

	// Проверяем, что ошибок нет и ID пользователя корректный
	assert.NoError(t, err)
	assert.Equal(t, int32(1), userID)
}

func TestBuyMerch(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockRepository{
		GetUserBalanceFunc: func(ctx context.Context, userID int32) (int32, error) {
			return 1000, nil // Баланс пользователя
		},
		GetMerchPriceFunc: func(ctx context.Context, merchID int32) (int32, error) {
			return 500, nil // Цена мерча
		},
		BuyMerchFunc: func(ctx context.Context, userID, merchID int32) error {
			return nil // Успешная покупка
		},
	}

	// Создаем сервис с мок-репозиторием
	coinService := service.NewCoinService(mockRepo)

	// Вызываем метод BuyMerch
	err := coinService.BuyMerch(context.Background(), 1, 1)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)
}

func TestTransferCoins(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockRepository{
		UserExistsFunc: func(ctx context.Context, username string) (db.UserExistsRow, error) {
			return db.UserExistsRow{ID: 2}, nil // Получатель существует
		},
		GetUserBalanceFunc: func(ctx context.Context, userID int32) (int32, error) {
			if userID == 1 {
				return 1000, nil // Баланс отправителя
			}
			return 500, nil // Баланс получателя
		},
		TransferCoinsFunc: func(ctx context.Context, fromUser, toUser, amount int32) error {
			return nil // Успешный перевод
		},
	}

	// Создаем сервис с мок-репозиторием
	coinService := service.NewCoinService(mockRepo)

	// Вызываем метод TransferCoins
	err := coinService.TransferCoins(context.Background(), 1, "testuser2", 200)

	// Проверяем, что ошибок нет
	assert.NoError(t, err)
}

func TestGetUserBalance(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockRepository{
		GetUserBalanceFunc: func(ctx context.Context, userID int32) (int32, error) {
			return 1000, nil // Баланс пользователя
		},
	}

	// Создаем сервис с мок-репозиторием
	coinService := service.NewCoinService(mockRepo)

	// Вызываем метод GetUserBalance
	infoResponse, err := coinService.GetUserBalance(context.Background(), 1)

	// Проверяем, что ошибок нет и баланс корректный
	assert.NoError(t, err)
	assert.Equal(t, 1000, *infoResponse.Coins)
}

func TestGetUserPurchases(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockRepository{
		GetUserPurchasesFunc: func(ctx context.Context, userID int32) ([]db.GetUserPurchasesRow, error) {
			return []db.GetUserPurchasesRow{
				{Name: "t-shirt"},
				{Name: "cup"},
			}, nil
		},
	}

	// Создаем сервис с мок-репозиторием
	coinService := service.NewCoinService(mockRepo)

	// Вызываем метод GetUserPurchases
	infoResponse, err := coinService.GetUserPurchases(context.Background(), 1)

	// Проверяем, что ошибок нет и инвентарь корректный
	assert.NoError(t, err)
	assert.Equal(t, 2, len(*infoResponse.Inventory))
}

func TestGetTransactions(t *testing.T) {
	// Создаем мок-репозиторий
	mockRepo := &MockRepository{
		GetTransactionsFunc: func(ctx context.Context, userID int32) ([]db.GetTransactionsRow, error) {
			return []db.GetTransactionsRow{
				{FromUser: sql.NullInt32{Int32: 2, Valid: true}, ToUser: sql.NullInt32{Int32: 1, Valid: true}, Amount: 100},
				{FromUser: sql.NullInt32{Int32: 1, Valid: true}, ToUser: sql.NullInt32{Int32: 3, Valid: true}, Amount: 50},
			}, nil
		},
	}

	// Создаем сервис с мок-репозиторием
	coinService := service.NewCoinService(mockRepo)

	// Вызываем метод GetTransactions
	infoResponse, err := coinService.GetTransactions(context.Background(), 1)

	// Проверяем, что ошибок нет и история транзакций корректна
	assert.NoError(t, err)
	assert.Equal(t, 1, len(*infoResponse.CoinHistory.Received))
	assert.Equal(t, 1, len(*infoResponse.CoinHistory.Sent))
}