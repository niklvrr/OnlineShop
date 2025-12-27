CREATE TABLE processed_messages (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(order_id, event_type)
);

CREATE INDEX idx_processed_messages_order_id ON processed_messages(order_id);

