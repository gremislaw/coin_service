# 💚 💚 💚 
# Coin Service 

Сервис для управления внутренней валютой (монетками) сотрудников. Реализован на Go с поддержкой REST API для покупки мерча, перевода монеток между сотрудниками и отслеживания транзакций. Сервис обеспечивает высокую производительность и отказоустойчивость, используя PostgreSQL в качестве основного хранилища данных.

## Описание

Сервис позволяет сотрудникам:

1. **Покупать мерч** — использовать монетки для приобретения товаров из внутреннего магазина.
2. **Переводить монетки** — отправлять монетки другим сотрудникам в знак благодарности или подарка.
3. **Отслеживать транзакции** — просматривать историю операций с монетками.

Каждый сотрудник при регистрации получает 1000 монеток. Баланс не может быть отрицательным, а все операции выполняются атомарно.

---

## Основные функции

### Поддерживаемые методы:

- **POST** `/api/auth`:
  - Регистрация или авторизация пользователя.
  - Пример запроса:
    ```bash
    curl -X POST http://localhost:8080/api/auth \
      -H "Content-Type: application/json" \
      -d '{"username": "user1", "password": "pass123"}'
    ```
  - Пример ответа:
    ```json
    {"token": "JWT_TOKEN"}
    ```

- **GET** `/api/buy/:merch_id`:
  - Покупка мерча по его ID.
  - Пример запроса:
    ```bash
    curl -X POST http://localhost:8080/api/buy/1 \
      -H "Authorization: Bearer JWT_TOKEN"
    ```
  - Пример ответа:
    ```json
    "merch purchased successfully"
    ```

- **POST** `/api/sendCoin`:
  - Перевод монеток другому сотруднику.
  - Пример запроса:
    ```bash
    curl -X POST http://localhost:8080/api/sencCoin \
      -H "Authorization: Bearer JWT_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"toUser": 2, "amount": 100}'
    ```
  - Пример ответа:
    ```json
    "coins transferred successfully"
    ```

- **GET** `/api/info`:
  - Получение текущего баланса пользователя.
  - Пример запроса:
    ```bash
    curl -X GET http://localhost:8080/api/info \
      -H "Authorization: Bearer JWT_TOKEN"
    ```
  - Пример ответа:
    ```json
    {"coinHistory":{"received":[{"amount":50,"fromUser":"user1"}],"sent":[{"amount":50,"toUser":"user3"},{"amount":50,"toUser":"user3"},  {"amount":50,"toUser":"user3"}]},"coins":820,"inventory":[{"quantity":1,"type":"t-shirt"}]}

    ```

---

## Стек технологий

- **Go** — основной язык программирования.
- **PostgreSQL** — база данных для хранения пользователей, транзакций и мерча.
- **JWT** — для аутентификации и авторизации.
- **Docker** — для упаковки и запуска сервиса.
- **Unit-тесты** — покрытие тестами с использованием `Go testing`.

---

## Установка и запуск

### Локальный запуск

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/your-username/avito-coin-service.git
   cd avito-coin-service
   ```

2. Запустите сервис:
   ```bash
   make run
   ```

   Сервис будет доступен по адресу: `http://localhost:8080`.

---

## Переменные окружения

- **DB_HOST** — хост для подключения к базе данных (по умолчанию `localhost`).
- **DB_PORT** — порт для подключения к базе данных (по умолчанию `5432`).
- **DB_USER** — имя пользователя для базы данных.
- **DB_PASSWORD** — пароль для базы данных.
- **DB_NAME** — имя базы данных.

---

## Настройка через Docker

Docker Compose используется для поднятия всех сервисов:

<details>
  <summary><strong>docker-compose.yml</strong></summary>

```yaml
services:
  postgres:
    image: postgres:latest
    restart: always
    environment:
      POSTGRES_USER: avito
      POSTGRES_PASSWORD: avito
      POSTGRES_DB: avito_coin
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data

  app:
    build: .
    container_name: avito_coin_service
    restart: on-failure
    depends_on: 
      - postgres
    environment:
      DB_HOST: "postgres"
      DB_PORT: "5432"
      DB_USER: "avito"
      DB_PASSWORD: "avito"
      DB_NAME: "avito_coin"
    ports:
      - "8080:8080"

volumes:
  db_data:
```
</details>

---

## Тестирование

Для внесения пользователей

Для тестирования сервиса используются unit-тесты. Чтобы запустить тесты, выполните:

```bash
make test
```

Код сервиса проверен линтерами (см. .golangci-lint)

Тесты покрывают следующие сценарии:
- Регистрация и авторизация пользователя.
- Покупка мерча.
- Перевод монеток между пользователями.
- Получение баланса и истории транзакций.

**Load Testing**

Запуск:
```bash
 make load_test
```

Результаты:

    Duration 3s

    Get User Balance Test Results:
    Requests: 1510
    Success Rate: 100.00%
    Latency (mean): 845.329062ms
    Latency (95th percentile): 2.428142488s
    Latency (99th percentile): 3.207231881s
    Bytes In (mean): 96.35
    Bytes Out (mean): 0.00

    Buy Merch Test Results:
    Requests: 1519
    Success Rate: 98.88%
    Latency (mean): 1.030131828s
    Latency (95th percentile): 2.859686845s
    Latency (99th percentile): 3.542084777s
    Bytes In (mean): 43.04
    Bytes Out (mean): 0.00

    Transfer Coins Test Results:
    Requests: 1510
    Success Rate: 99.54%
    Latency (mean): 1.196866201s
    Latency (95th percentile): 3.041751118s
    Latency (99th percentile): 3.7649174s
    Bytes In (mean): 44.97
    Bytes Out (mean): 28.88

---

## Архитектура и структура проекта

Проект состоит из слоев:
- **Handler** — обработка HTTP-запросов, валидация, логирование.
- **Service** — бизнес-логика: покупка мерча, перевод монет, управление балансом.
- **Repository** — работа с базой данных: создание пользователей, мерча, транзакций.

Структура проекта:
```
├── cmd
│   └── main.go
├── internal
│   ├── config
│   ├── handler
│   ├── repository
│   ├── service
│   └── db
├── Dockerfile
├── docker-compose.yml
├── README.md
└── go.mod
```

---

## Проблемы

Сервис не мог удержать высокую нагрузку в 1000rps 

Я предположил, что 1000 горутин скапливались из-за ограничения бд и увеличил этот порог

Решением было увеличение максимального количества коннектов к бд до 1000 в ручную в конфигах и задать в сервисе ограничители коннектов

---
> **"Не удерживай то, что уходит, и не отвергай то, что приходит."**  
> — Кодо Саваки  
