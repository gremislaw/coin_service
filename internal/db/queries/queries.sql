-- name: CreateUser :one
INSERT INTO users (username, password)
VALUES ($1, $2)
RETURNING id;

-- name: CreateMerch :exec
INSERT INTO merch (name, price)
VALUES ($1, $2);

-- name: UserExists :one
SELECT id, password
FROM users
WHERE username = $1;

-- name: GetUserBalance :one
SELECT balance 
FROM users 
WHERE id = $1;

-- name: GetMerchPrice :one
SELECT price 
FROM merch 
WHERE id = $1;

-- name: UpdateUserBalance :exec
UPDATE users
SET balance = $1
WHERE id = $2;

-- name: TransferCoins :exec
-- Перевод монет от одного пользователя к другому
INSERT INTO transactions (from_user, to_user, amount)
VALUES ($1, $2, $3);

-- name: BuyMerch :exec
-- Покупка товара пользователем
INSERT INTO purchases (user_id, merch_id)
VALUES ($1, $2);

-- name: GetUserPurchases :many
-- Получение списка всех покупок пользователя
SELECT m.name, p.purchase_time 
FROM purchases p
JOIN merch m ON p.merch_id = m.id
WHERE p.user_id = $1
ORDER BY p.purchase_time DESC;

-- name: GetTransactions :many
-- Получение списка транзакций для пользователя (кто кому передавал монеты и в каком количестве)
SELECT t.from_user, t.to_user, t.amount, t.transaction_time
FROM transactions t
WHERE t.from_user = $1 OR t.to_user = $1
ORDER BY t.transaction_time DESC;