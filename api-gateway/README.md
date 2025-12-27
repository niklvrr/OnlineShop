# API Gateway

API Gateway предоставляет REST API для доступа к микросервисам.

## Запуск

### Переменные окружения

Создайте `.env` файл:

```env
HTTP_PORT=8080
ORDER_SERVICE_URL=localhost:50051
PAYMENT_SERVICE_URL=localhost:50052
```

### Запуск сервиса

```bash
go run cmd/server/main.go
```

## REST API

### Orders

- `POST /api/v1/orders` - создание заказа
- `GET /api/v1/orders/{order_id}` - получение статуса заказа

### Payments

- `POST /api/v1/payments` - инициация платежа
- `GET /api/v1/payments/{payment_id}` - получение статуса платежа

