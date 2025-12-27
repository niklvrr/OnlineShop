CREATE TYPE message_type AS ENUM('PENDING', 'SENT');

CREATE TABLE outbox (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status message_type NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    sent_at TIMESTAMP,
    attempts INT DEFAULT 0,
    unique_key VARCHAR(255)
);

CREATE INDEX idx_outbox_status_pending ON outbox(status, created_at)
    WHERE status = 'PENDING';
CREATE INDEX idx_outbox_aggregate_id ON outbox(aggregate_id);
CREATE UNIQUE INDEX idx_outbox_unique_key ON outbox(unique_key) WHERE unique_key IS NOT NULL;

