# Payment Service

Payment Service обрабатывает платежи и управляет счетами пользователей.

## Запуск

### Переменные окружения

Создайте `.env` файл:

```env
APP_ENV_LEVEL=debug
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment-service
KAFKA_BROKERS=localhost:9092
GRPC_PORT=50052
```

### Миграции

```bash
migrate -path ./migrations -database "postgresql://postgres:postgres@localhost:5432/payment-service?sslmode=disable" up
```

### Запуск сервиса

```bash
go run cmd/server/main.go
```

## gRPC API

- `CreateAccount` - создание счета для пользователя
- `ReplenishAccount` - пополнение счета
- `GetBalance` - получение баланса
- `DebitToAccount` - списание со счета
- `GetAccountByUserId` - получение счета по user_id
- `InitiatePayment` - инициация платежа
- `GetPaymentStatus` - получение статуса платежа

## Kafka

Сервис подписывается на топик `order-events` и обрабатывает события `OrderCreated`.

