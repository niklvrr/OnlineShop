CREATE TYPE payment_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED');

CREATE TABLE payments (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    account_id UUID NOT NULL,
    amount FLOAT NOT NULL,
    payment_method VARCHAR(255) NOT NULL,
    status payment_status DEFAULT 'PENDING',
    idempotency_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE UNIQUE INDEX idx_payments_idempotency ON payments(order_id, idempotency_key);
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);

