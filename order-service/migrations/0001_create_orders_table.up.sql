CREATE TYPE order_status AS ENUM ('NEW', 'FINISHED', 'CANCELED');

CREATE TABLE orders (
    id UUID NOT NULL PRIMARY KEY,
    user_id UUID NOT NULL,
    amount FLOAT DEFAULT 0.0,
    status order_status DEFAULT 'NEW',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);