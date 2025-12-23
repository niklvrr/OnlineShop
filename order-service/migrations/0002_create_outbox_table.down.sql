DROP INDEX IF EXISTS idx_outbox_aggregate_id;
DROP INDEX IF EXISTS idx_outbox_status_pending;

DROP TABLE IF EXISTS outbox;

DROP TYPE IF EXISTS message_type;