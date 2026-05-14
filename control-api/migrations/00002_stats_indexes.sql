-- +goose Up
-- composite index to speed up dashboard stats queries that filter by time range + payment columns
CREATE INDEX idx_request_logs_created_payment
    ON request_logs (created_at, payment_required, payment_completed);

-- +goose Down
DROP INDEX IF EXISTS idx_request_logs_created_payment;
