-- +goose Up
-- +goose StatementBegin

CREATE TABLE payment_methods (
    id             UUID PRIMARY KEY,
    code           VARCHAR NOT NULL UNIQUE,
    protocol       VARCHAR NOT NULL CHECK (protocol IN ('x402', 'mpp')),
    name           VARCHAR NOT NULL,
    caip2_chain_id VARCHAR,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payment_method_assets (
    id                UUID PRIMARY KEY,
    payment_method_id UUID NOT NULL REFERENCES payment_methods (id),
    symbol            VARCHAR NOT NULL,
    contract_address  VARCHAR,
    decimals          INTEGER NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (payment_method_id, symbol)
);

CREATE TABLE facilitators (
    id         UUID PRIMARY KEY,
    name       VARCHAR NOT NULL,
    url        VARCHAR NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE project_payment_methods (
    id                UUID PRIMARY KEY,
    project_id        UUID NOT NULL REFERENCES projects (id),
    payment_method_id UUID NOT NULL REFERENCES payment_methods (id),
    asset_id          UUID NOT NULL REFERENCES payment_method_assets (id),
    scheme            VARCHAR NOT NULL,
    facilitator_id    UUID NOT NULL REFERENCES facilitators (id),
    payout_address    VARCHAR,
    config            JSONB,
    enabled           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (project_id, payment_method_id, asset_id, scheme)
);

DROP TABLE IF EXISTS project_payment_configs;
DROP TABLE IF EXISTS payment_channel_assets CASCADE;
DROP TABLE IF EXISTS payment_channels CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS project_payment_methods;
DROP TABLE IF EXISTS facilitators;
DROP TABLE IF EXISTS payment_method_assets;
DROP TABLE IF EXISTS payment_methods;

CREATE TABLE payment_channels (
    id         UUID PRIMARY KEY,
    protocol   VARCHAR NOT NULL,
    method     VARCHAR NOT NULL,
    scheme     VARCHAR NOT NULL,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (protocol, method, scheme)
);

CREATE TABLE payment_channel_assets (
    id                 UUID PRIMARY KEY,
    payment_channel_id UUID NOT NULL REFERENCES payment_channels (id),
    asset_symbol       VARCHAR NOT NULL,
    asset_address      VARCHAR,
    decimals           INTEGER,
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (payment_channel_id, asset_symbol)
);

CREATE TABLE project_payment_configs (
    id                       UUID PRIMARY KEY,
    project_id               UUID NOT NULL REFERENCES projects (id),
    payment_channel_id       UUID NOT NULL REFERENCES payment_channels (id),
    payment_channel_asset_id UUID REFERENCES payment_channel_assets (id),
    name                     VARCHAR NOT NULL,
    payout_address           VARCHAR,
    enabled                  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at               TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd
