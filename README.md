# Online Shop - Микросервисная архитектура

Проект реализует микросервисную архитектуру для интернет-магазина с использованием Go, gRPC, Kafka и PostgreSQL.

## Архитектура

- **Order Service** - управление заказами
- **Payment Service** - обработка платежей
- **API Gateway** - REST API для клиентов
- **Frontend** - минимальный UI

## Быстрый старт

### Запуск через Docker Compose

```bash
docker-compose up --build
```

Это запустит:
- PostgreSQL (2 инстанса для order и payment)
- Zookeeper и Kafka
- Order Service (gRPC порт 50051)
- Payment Service (gRPC порт 50052)
- API Gateway (HTTP порт 8080)
- Frontend (HTTP порт 80)

### Применение миграций

После запуска контейнеров примените миграции:

```bash
# Order Service
migrate -path ./order-service/migrations -database "postgresql://postgres:postgres@localhost:5432/order-service?sslmode=disable" up

# Payment Service
migrate -path ./payment-service/migrations -database "postgresql://postgres:postgres@localhost:5433/payment-service?sslmode=disable" up
```

### Проверка работы

1. Откройте браузер: http://localhost
2. Создайте заказ через UI
3. Проверьте статус заказа
4. Инициируйте платеж

## API Endpoints

### Orders

- `POST /api/v1/orders` - создание заказа
- `GET /api/v1/orders/{order_id}` - получение статуса заказа

### Payments

- `POST /api/v1/payments` - инициация платежа
- `GET /api/v1/payments/{payment_id}` - получение статуса платежа

## Примеры запросов

### Создание заказа

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-1",
    "items": [
      {"sku": "sku-1", "qty": 1, "price": 100}
    ],
    "metadata": {}
  }'
```

### Получение статуса заказа

```bash
curl http://localhost:8080/api/v1/orders/{order_id}
```

### Инициация платежа

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "order-id",
    "user_id": "user-1",
    "amount": 100,
    "payment_method": "card"
  }'
```

## Структура проекта

```
.
├── order-service/       # Order Service
├── payment-service/     # Payment Service
├── api-gateway/         # API Gateway
├── frontend/            # Frontend
├── proto-contracts/     # gRPC контракты
├── common_library/      # Общая библиотека
└── docker-compose.yml   # Docker Compose конфигурация
```

## Технологии

- **Backend**: Go 1.25
- **База данных**: PostgreSQL
- **Брокер сообщений**: Apache Kafka
- **gRPC**: для межсервисного взаимодействия
- **REST**: для клиентского API

## Документация

Подробная документация находится в:
- `REPORT.md` - отчет о выполненной работе
- `openapi.yaml` - OpenAPI спецификация
- `postman_collection.json` - Postman коллекция

