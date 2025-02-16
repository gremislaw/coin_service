-- +goose Up

-- Создание таблицы пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- Пароль (строка, для хранения пароля)
    balance INT NOT NULL DEFAULT 1000 -- Изначально каждому сотруднику начисляется 1000 монет
);

-- Создание таблицы мерча
CREATE TABLE merch (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL -- Цена товара
);

-- Создание таблицы покупок (связь между пользователями и товарами)
CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    merch_id INT REFERENCES merch(id),
    purchase_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы транзакций (для передачи монет между пользователями)
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user INT REFERENCES users(id),
    to_user INT REFERENCES users(id),
    amount INT NOT NULL,
    transaction_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Индекс для пользователей по имени пользователя (для быстрого поиска по имени)
CREATE INDEX IF NOT EXISTS idx_users_username
ON users (username);

-- Индекс для транзакций по идентификатору пользователя (для быстрого поиска всех транзакций пользователя)
CREATE INDEX IF NOT EXISTS idx_transactions_user_time
ON transactions (from_user, to_user, transaction_time);

-- Индекс для покупок по идентификатору пользователя (для быстрого поиска всех покупок пользователя)
CREATE INDEX IF NOT EXISTS idx_purchases_user_id
ON purchases (user_id);

-- Индекс для товаров по названию (для быстрого поиска товаров по названию)
CREATE INDEX IF NOT EXISTS idx_merch_name
ON merch (name);

-- Индекс для транзакций по дате (для быстрого поиска транзакций по дате)
CREATE INDEX IF NOT EXISTS idx_transactions_time
ON transactions (transaction_time);

-- Индекс для покупок по товару (для быстрого поиска всех покупок по конкретному товару)
CREATE INDEX IF NOT EXISTS idx_purchases_merch_id
ON purchases (merch_id);

-- +goose Down

-- Удаление таблицы покупок
DROP TABLE IF EXISTS purchases;

-- Удаление таблицы пользователей
DROP TABLE IF EXISTS users;

-- Удаление таблицы товаров
DROP TABLE IF EXISTS merch;

-- Удаление таблицы транзакций
DROP TABLE IF EXISTS transactions;

-- Удаление индексов (если они были созданы)
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_purchases_user_id;

