package repository

import (
	"context"
	"database/sql"
	"fmt"

	"avito_coin/internal/db"
)

// Repository - интерфейс репозитория для операций с монетками и мерчем.
type Repository interface {
	CreateUser(ctx context.Context, username, password string) (int32, error)
	CreateMerch(ctx context.Context, name string, price int32) error
	BuyMerch(ctx context.Context, userID, merchID int32) error
	GetMerchPrice(ctx context.Context, merchID int32) (int32, error)
	TransferCoins(ctx context.Context, fromUser, toUser, amount int32) error
	GetUserBalance(ctx context.Context, userID int32) (int32, error)
	GetUserPurchases(ctx context.Context, userID int32) ([]db.GetUserPurchasesRow, error)
	GetTransactions(ctx context.Context, userID int32) ([]db.GetTransactionsRow, error)
	UpdateUserBalance(ctx context.Context, userID int32, balance int32) error
	UserExists(ctx context.Context, username string) (db.UserExistsRow, error)
}

// coinRepository - структура, которая реализует интерфейс Repository.
type coinRepository struct {
	queries *db.Queries
	db      *sql.DB
}

// NewRepository - функция для создания нового репозитория.
func NewRepository(database *sql.DB) Repository {
	return &coinRepository{
		queries: db.New(database),
		db:      database,
	}
}

// CreateUser - создание пользователя с именем и паролем.
func (r *coinRepository) CreateUser(ctx context.Context, username, password string) (int32, error) {
	return r.queries.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Password: password,
	})
}

// CreateMerch - создание мерча с именем и ценой.
func (r *coinRepository) CreateMerch(ctx context.Context, name string, price int32) error {
	return r.queries.CreateMerch(ctx, db.CreateMerchParams{
		Name:  name,
		Price: price,
	})
}

// BuyMerch - покупка мерча пользователем.
func (r *coinRepository) BuyMerch(ctx context.Context, userID, merchID int32) error {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Откат транзакции в случае ошибки
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Создаём новый экземпляр queries для работы в транзакции
	qtx := r.queries.WithTx(tx)

	// Проверяем, достаточно ли монеток на балансе для покупки
	balance, err := qtx.GetUserBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("error retrieving user balance: %w", err)
	}

	// Получение цены мерча
	price, err := qtx.GetMerchPrice(ctx, merchID)

	// Есть ли достаточное количество монет для покупки
	if balance < price {
		return fmt.Errorf("insufficient balance for purchase")
	}

	// Выполняем покупку
	err = qtx.BuyMerch(ctx, db.BuyMerchParams{
		UserID:  sql.NullInt32{Int32: userID, Valid: true},
		MerchID: sql.NullInt32{Int32: merchID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error buying merch: %w", err)
	}

	// Обновляем баланс пользователя
	newBalance := balance - price

	err = qtx.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		Balance: newBalance,
		ID:      userID,
	})
	if err != nil {
		return fmt.Errorf("error updating user balance after merch purchase: %w", err)
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// TransferCoins - перевод монет от одного пользователя к другому.
func (r *coinRepository) TransferCoins(ctx context.Context, fromUser, toUser, amount int32) error {
	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Откат транзакции в случае ошибки
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Создаём новый экземпляр queries для работы в транзакции
	qtx := r.queries.WithTx(tx)

	// Проверяем, достаточно ли монет у отправителя
	balance, err := qtx.GetUserBalance(ctx, fromUser)
	if err != nil {
		return fmt.Errorf("error retrieving user balance: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("insufficient balance to transfer")
	}

	// Выполняем перевод монет
	err = qtx.TransferCoins(ctx, db.TransferCoinsParams{
		FromUser: sql.NullInt32{Int32: fromUser, Valid: true},
		ToUser:   sql.NullInt32{Int32: toUser, Valid: true},
		Amount:   amount,
	})
	if err != nil {
		return fmt.Errorf("error transferring coins: %w", err)
	}

	// Обновляем балансы обоих пользователей
	err = qtx.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		Balance: balance - amount,
		ID:      fromUser,
	})
	if err != nil {
		return fmt.Errorf("error updating sender balance: %w", err)
	}

	// Получаем баланс получателя и обновляем его
	receiverBalance, err := qtx.GetUserBalance(ctx, toUser)
	if err != nil {
		return fmt.Errorf("error retrieving receiver balance: %w", err)
	}

	err = qtx.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		Balance: receiverBalance + amount,
		ID:      toUser,
	})
	if err != nil {
		return fmt.Errorf("error updating receiver balance: %w", err)
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// GetUserPurchases - получение всех покупок пользователя.
func (r *coinRepository) GetUserPurchases(ctx context.Context, userID int32) ([]db.GetUserPurchasesRow, error) {
	return r.queries.GetUserPurchases(ctx, sql.NullInt32{Int32: userID, Valid: true})
}

// GetTransactions - получение списка всех транзакций пользователя.
func (r *coinRepository) GetTransactions(ctx context.Context, userID int32) ([]db.GetTransactionsRow, error) {
	return r.queries.GetTransactions(ctx, sql.NullInt32{Int32: userID, Valid: true})
}

// UpdateUserBalance - обновление баланса пользователя.
func (r *coinRepository) UpdateUserBalance(ctx context.Context, userID int32, balance int32) error {
	return r.queries.UpdateUserBalance(ctx, db.UpdateUserBalanceParams{
		Balance: balance,
		ID:      userID,
	})
}

// GetUserBalance - получение баланса пользователя.
func (r *coinRepository) GetUserBalance(ctx context.Context, userID int32) (int32, error) {
	return r.queries.GetUserBalance(ctx, userID)
}

// UserExists - существует ли пользователь.
func (r *coinRepository) UserExists(ctx context.Context, username string) (db.UserExistsRow, error) {
	return r.queries.UserExists(ctx, username)
}

// GetUserBalance - получение баланса пользователя.
func (r *coinRepository) GetMerchPrice(ctx context.Context, merchID int32) (int32, error) {
	return r.queries.GetMerchPrice(ctx, merchID)
}
