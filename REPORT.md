# Отчет о выполненной работе

## Выполненные требования

### 1. Сервисная разделённость ✅
- OrderService реализован как отдельное приложение с Dockerfile и README
- PaymentService реализован как отдельное приложение с Dockerfile и README
- API Gateway реализован как отдельное приложение с Dockerfile и README
- Каждый сервис находится в отдельной папке с собственной структурой

**Проверка**: В корне проекта есть папки `order-service/`, `payment-service/`, `api-gateway/`, каждая содержит Dockerfile и README.md

### 2. База данных и миграции ✅
- Для OrderService созданы миграции в формате go-migrate:
  - `0001_create_orders_table.up.sql` / `.down.sql`
  - `0002_create_outbox_table.up.sql` / `.down.sql`
  - `0003_create_order_items_table.up.sql` / `.down.sql`
- Для PaymentService созданы миграции:
  - `0001_create_accounts_table.up.sql` / `.down.sql`
  - `0002_create_payments_table.up.sql` / `.down.sql`
  - `0003_create_outbox_table.up.sql` / `.down.sql`
  - `0004_create_processed_messages_table.up.sql` / `.down.sql`

**Проверка**:
```bash
migrate -path ./order-service/migrations -database "postgresql://postgres:postgres@localhost:5432/order-service?sslmode=disable" up
migrate -path ./payment-service/migrations -database "postgresql://postgres:postgres@localhost:5433/payment-service?sslmode=disable" up
```

### 3. Использование database/sql ✅
- Весь доступ к PostgreSQL реализован через `database/sql`
- Используется драйвер `github.com/jackc/pgx/v5/stdlib`
- ORM (GORM) не используется

**Проверка**: В коде нет импортов GORM, используются только `database/sql` и `sqlx` не используется

### 4. gRPC контракты и реализация ✅
- Proto-файлы обновлены:
  - `order-service.proto` - добавлена поддержка items в CreateOrderRequest
  - `payment-service.proto` - добавлены методы InitiatePayment и GetPaymentStatus
- gRPC серверы реализованы в обоих сервисах
- Все методы из proto реализованы

**Проверка**: 
- Proto файлы находятся в `proto-contracts/`
- Сгенерированные stubs в `proto-contracts/gen/go/`
- Для генерации: `cd proto-contracts && make proto`

### 5. Kafka используется правильно ✅
- OrderService публикует события через outbox pattern
- Реализован dispatcher, который читает из outbox и публикует в Kafka
- PaymentService подписывается на топик `order-events` и обрабатывает события
- Используется idempotent producer

**Проверка**:
- Dispatcher в `order-service/internal/infrastructure/kafka/dispatcher.go`
- Consumer в `payment-service/internal/infrastructure/kafka/consumer.go`
- Outbox таблица в миграциях

### 6. Transactional Outbox + Idempotency ✅
- Outbox реализован в обоих сервисах
- Dispatcher помечает сообщения как отправленные после успешной публикации
- PaymentService проверяет дубликаты через таблицу `processed_messages`
- Idempotency реализована через уникальный ключ (order_id + idempotency_key)

**Проверка**:
- Outbox таблицы в миграциях
- Dispatcher обновляет статус сообщений
- PaymentService проверяет `processed_messages` перед обработкой

### 7. API Gateway - REST -> gRPC routing ✅
- Все REST endpoints корректно маршрутизируют в gRPC методы
- Gateway не реализует бизнес-логику, только маршрутизацию
- Реализованы все требуемые endpoints:
  - POST /api/v1/orders
  - GET /api/v1/orders/{order_id}
  - POST /api/v1/payments
  - GET /api/v1/payments/{payment_id}

**Проверка**:
- OpenAPI спецификация в `openapi.yaml`
- Примеры curl в README.md

### 8. Фронтенд ✅
- Минимальный UI реализован
- Запускается в отдельном контейнере
- Может создать заказ и посмотреть статус
- Может инициировать платеж и посмотреть статус

**Проверка**:
```bash
docker-compose up frontend
# Открыть http://localhost
```

### 9. Документация ✅
- README.md для каждого сервиса
- Общий README.md в корне
- OpenAPI спецификация (openapi.yaml)
- Postman коллекция (postman_collection.json)

**Проверка**: Все файлы находятся в репозитории

### 10. Контроль качества кода ✅
- Код отформатирован (gofmt)
- Нет лишних зависимостей
- Названия переменных и структура следуют стилю существующего OrderService
- Код без комментариев (как просили)

**Проверка**: `gofmt -l` не показывает файлов

## Структура проекта

```
.
├── order-service/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/
│   │   ├── handler/
│   │   ├── infrastructure/
│   │   │   ├── kafka/
│   │   │   └── pgdb/
│   │   └── usecase/
│   ├── migrations/
│   ├── Dockerfile
│   └── README.md
├── payment-service/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/
│   │   ├── handler/
│   │   ├── infrastructure/
│   │   │   ├── kafka/
│   │   │   └── pgdb/
│   │   └── usecase/
│   ├── migrations/
│   ├── Dockerfile
│   └── README.md
├── api-gateway/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── client/
│   │   ├── config/
│   │   └── handler/
│   ├── Dockerfile
│   └── README.md
├── frontend/
│   ├── index.html
│   ├── Dockerfile
│   └── README.md
├── proto-contracts/
│   ├── order/order-service.proto
│   ├── payment/payment-service.proto
│   └── gen/go/
├── docker-compose.yml
├── openapi.yaml
├── postman_collection.json
├── README.md
└── REPORT.md
```

## Как проверить

### 1. Запуск всего стека

```bash
docker-compose up --build
```

### 2. Применение миграций

```bash
# Order Service
migrate -path ./order-service/migrations -database "postgresql://postgres:postgres@localhost:5432/order-service?sslmode=disable" up

# Payment Service
migrate -path ./payment-service/migrations -database "postgresql://postgres:postgres@localhost:5433/payment-service?sslmode=disable" up
```

### 3. Создание заказа

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

### 4. Проверка статуса заказа

```bash
curl http://localhost:8080/api/v1/orders/{order_id}
```

### 5. Инициация платежа

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "{order_id}",
    "user_id": "user-1",
    "amount": 100,
    "payment_method": "card"
  }'
```

### 6. Проверка статуса платежа

```bash
curl http://localhost:8080/api/v1/payments/{payment_id}
```

### 7. Проверка outbox

```sql
SELECT * FROM outbox WHERE status = 'PENDING';
SELECT * FROM outbox WHERE status = 'SENT';
```

### 8. Проверка Kafka

```bash
docker exec -it <kafka-container> kafka-console-consumer --bootstrap-server localhost:9092 --topic order-events --from-beginning
```

## Изменения в proto-контрактах

1. **order-service.proto**:
   - Добавлен `OrderItem` message
   - `CreateOrderRequest` теперь содержит `items` и `metadata`
   - `GetOrderStatusResponse` расширен полями `order_id`, `amount`, `user_id`

2. **payment-service.proto**:
   - Добавлены методы `InitiatePayment` и `GetPaymentStatus`
   - Добавлены соответствующие request/response messages

## Особенности реализации

1. **Idempotency**: Реализована через уникальный ключ в таблице payments (order_id + idempotency_key)
2. **Outbox Pattern**: Используется для гарантированной доставки событий в Kafka
3. **Transactional Outbox**: События записываются в outbox в той же транзакции, что и основная операция
4. **Kafka Consumer**: PaymentService обрабатывает события OrderCreated и создает счета для пользователей
5. **Error Handling**: Все ошибки обрабатываются и логируются

## Заключение

Все требования ТЗ выполнены. Система готова к использованию и соответствует всем acceptance criteria.

