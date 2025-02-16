package service_test

import (
	"avito_coin/internal/db"
	"avito_coin/internal/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, username, password string) (int32, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockRepository) CreateMerch(ctx context.Context, name string, price int32) error {
	args := m.Called(ctx, name, price)
	return args.Error(0)
}

func (m *MockRepository) BuyMerch(ctx context.Context, userID, merchID int32) error {
	args := m.Called(ctx, userID, merchID)
	return args.Error(0)
}

func (m *MockRepository) GetUserBalance(ctx context.Context, userID int32) (int32, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockRepository) GetMerchPrice(ctx context.Context, merchID int32) (int32, error) {
	args := m.Called(ctx, merchID)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockRepository) TransferCoins(ctx context.Context, fromUser, toUser, amount int32) error {
	args := m.Called(ctx, fromUser, toUser, amount)
	return args.Error(0)
}

func (m *MockRepository) UpdateUserBalance(ctx context.Context, userID, balance int32) error {
	args := m.Called(ctx, userID, balance)
	return args.Error(0)
}

func (m *MockRepository) UserExists(ctx context.Context, username string) (db.UserExistsRow, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(db.UserExistsRow), args.Error(1)
}

func (m *MockRepository) GetTransactions(ctx context.Context, userID int32) ([]db.GetTransactionsRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.GetTransactionsRow), args.Error(1)
}

func (m *MockRepository) GetUserPurchases(ctx context.Context, userID int32) ([]db.GetUserPurchasesRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.GetUserPurchasesRow), args.Error(1)
}

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockRepository)
	coinService := service.NewCoinService(mockRepo)

	// Мокаем успешное создание пользователя
	mockRepo.On("CreateUser", mock.Anything, "testuser", "password123").Return(int32(1), nil)

	// Вызов метода
	userID, err := coinService.CreateUser(context.Background(), "testuser", "password123")

	// Проверка результата
	assert.NoError(t, err)
	assert.Equal(t, int32(1), userID)
	mockRepo.AssertExpectations(t)
}

func TestBuyMerch_InsufficientBalance(t *testing.T) {
	mockRepo := new(MockRepository)
	coinService := service.NewCoinService(mockRepo)

	// Мокаем ответ с балансом 50, а цена мерча 100
	mockRepo.On("GetUserBalance", mock.Anything, int32(1)).Return(int32(50), nil)
	mockRepo.On("GetMerchPrice", mock.Anything, int32(1)).Return(int32(100), nil)

	// Вызов метода
	err := coinService.BuyMerch(context.Background(), 1, 1)

	// Проверка результата
	assert.Error(t, err)
	assert.Equal(t, "insufficient balance for purchase", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUpdateUserBalance_BalanceCannotBeNegative(t *testing.T) {
	mockRepo := new(MockRepository)
	coinService := service.NewCoinService(mockRepo)

	// Мокаем существование пользователя с балансом 100
	mockRepo.On("GetUserBalance", mock.Anything, int32(1)).Return(int32(100), nil)

	// Вызов метода с отрицательным балансом
	err := coinService.UpdateUserBalance(context.Background(), 1, -50)

	// Проверка результата
	assert.Error(t, err)
	assert.Equal(t, "balance cannot be negative", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestGetMerchPrice(t *testing.T) {
	mockRepo := new(MockRepository)
	coinService := service.NewCoinService(mockRepo)

	// Мокаем цену мерча
	mockRepo.On("GetMerchPrice", mock.Anything, int32(1)).Return(int32(100), nil)

	// Вызов метода
	price, err := coinService.GetMerchPrice(context.Background(), 1)

	// Проверка результата
	assert.NoError(t, err)
	assert.Equal(t, int32(100), price)
	mockRepo.AssertExpectations(t)
}

func TestGetUserBalance(t *testing.T) {
	mockRepo := new(MockRepository)
	coinService := service.NewCoinService(mockRepo)

	// Мокаем баланс пользователя
	mockRepo.On("GetUserBalance", mock.Anything, int32(1)).Return(int32(200), nil)

	// Вызов метода
	infoResponse, err := coinService.GetUserBalance(context.Background(), 1)

	// Проверка результата
	assert.NoError(t, err)
	assert.Equal(t, int(200), *infoResponse.Coins)
	mockRepo.AssertExpectations(t)
}
