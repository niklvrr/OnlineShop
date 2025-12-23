CREATE TYPE message_type AS ENUM('PENDING', 'SENT');

CREATE TABLE outbox (
    id UUID NOT NULL PRIMARY KEY,
    order_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL ,
    payload JSONB NOT NULL ,
    status message_type NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_status_pending ON outbox(status, created_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_aggregate_id ON outbox(aggregate_id);