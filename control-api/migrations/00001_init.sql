-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR NOT NULL UNIQUE,
    password_hash VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE projects (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL REFERENCES users (id),
    name VARCHAR NOT NULL,
    slug VARCHAR NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment_channels (
    id UUID PRIMARY KEY,
    protocol VARCHAR NOT NULL,
    method VARCHAR NOT NULL,
    scheme VARCHAR NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (protocol, method, scheme)
);

CREATE TABLE payment_channel_assets (
    id UUID PRIMARY KEY,
    payment_channel_id UUID NOT NULL REFERENCES payment_channels (id),
    asset_symbol VARCHAR NOT NULL,
    asset_address VARCHAR,
    decimals INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (payment_channel_id, asset_symbol)
);

CREATE TABLE project_payment_configs (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id),
    payment_channel_id UUID NOT NULL REFERENCES payment_channels (id),
    payment_channel_asset_id UUID REFERENCES payment_channel_assets (id),
    name VARCHAR NOT NULL,
    payout_address VARCHAR,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE project_routes_settings (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) UNIQUE,
    base_url VARCHAR NOT NULL,
    auth_header_name VARCHAR,
    auth_header_value VARCHAR,
    allow_unmatched BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE routes (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id),
    name VARCHAR NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    path_pattern VARCHAR NOT NULL,
    price_amount INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    free BOOLEAN NOT NULL DEFAULT FALSE,
    price_usd TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE request_logs (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id),
    outbound_route_id UUID REFERENCES routes (id),
    request_id VARCHAR NOT NULL UNIQUE,
    method VARCHAR NOT NULL,
    path VARCHAR NOT NULL,
    client_ip VARCHAR,
    user_agent TEXT,
    status VARCHAR NOT NULL CHECK (
        status IN (
            'started',
            'payment_required',
            'payment_completed',
            'proxying',
            'completed',
            'failed',
            'upstream_timeout',
            'upstream_error'
        )
    ),
    payment_required BOOLEAN NOT NULL DEFAULT FALSE,
    payment_requested_at TIMESTAMP,
    payment_completed BOOLEAN NOT NULL DEFAULT FALSE,
    payment_completed_at TIMESTAMP,
    payment_channel_id UUID REFERENCES payment_channels (id),
    payment_channel_asset_id UUID REFERENCES payment_channel_assets (id),
    amount_paid BIGINT,
    amount_usd NUMERIC(18, 8),
    upstream_url TEXT,
    upstream_status_code INTEGER,
    upstream_response_time_ms INTEGER,
    final_status_code INTEGER,
    error_type VARCHAR,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_request_logs_created_at ON request_logs (created_at);
CREATE INDEX idx_request_logs_status ON request_logs (status);
CREATE INDEX idx_request_logs_payment_completed ON request_logs (payment_completed);

CREATE TABLE request_events (
    id UUID PRIMARY KEY,
    request_log_id UUID NOT NULL REFERENCES request_logs (id) ON DELETE CASCADE,
    event_type VARCHAR NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_request_events_request_log_id ON request_events (request_log_id);
CREATE INDEX idx_request_events_event_type ON request_events (event_type);
CREATE INDEX idx_request_events_created_at ON request_events (created_at);

CREATE TABLE project_daily_stats (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id),
    date DATE NOT NULL,
    requests_count BIGINT NOT NULL DEFAULT 0,
    paid_requests_count BIGINT NOT NULL DEFAULT 0,
    revenue_usd NUMERIC(18, 8) NOT NULL DEFAULT 0,
    upstream_errors_count BIGINT NOT NULL DEFAULT 0,
    upstream_timeouts_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, date)
);

CREATE INDEX idx_project_daily_stats_project_id ON project_daily_stats (project_id);
CREATE INDEX idx_project_daily_stats_date ON project_daily_stats (date);

CREATE TABLE route_daily_stats (
    id UUID PRIMARY KEY,
    outbound_route_id UUID NOT NULL REFERENCES routes (id),
    date DATE NOT NULL,
    requests_count BIGINT NOT NULL DEFAULT 0,
    paid_requests_count BIGINT NOT NULL DEFAULT 0,
    revenue_usd NUMERIC(18, 8) NOT NULL DEFAULT 0,
    success_responses_count BIGINT NOT NULL DEFAULT 0,
    error_responses_count BIGINT NOT NULL DEFAULT 0,
    avg_response_time_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (outbound_route_id, date)
);

CREATE INDEX idx_route_daily_stats_date ON route_daily_stats (date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS route_daily_stats;
DROP TABLE IF EXISTS project_daily_stats;
DROP TABLE IF EXISTS request_events;
DROP TABLE IF EXISTS request_logs;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS project_routes_settings;
DROP TABLE IF EXISTS project_payment_configs;
DROP TABLE IF EXISTS payment_channel_assets;
DROP TABLE IF EXISTS payment_channels;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
