# Order Service

Order Service управляет заказами пользователей.

## Запуск

### Переменные окружения

Создайте `.env` файл:

```env
APP_ENV_LEVEL=debug
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=order-service
KAFKA_BROKERS=localhost:9092
GRPC_PORT=50051
```

### Миграции

```bash
migrate -path ./migrations -database "postgresql://postgres:postgres@localhost:5432/order-service?sslmode=disable" up
```

### Запуск сервиса

```bash
go run cmd/server/main.go
```

## gRPC API

- `CreateOrder` - создание заказа
- `GetAllOrders` - получение всех заказов
- `GetOrderStatus` - получение статуса заказа

## Kafka

Сервис публикует события в топик `order-events` через outbox pattern.

