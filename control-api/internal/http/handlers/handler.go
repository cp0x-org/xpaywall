package handlers

import (
	"github.com/cp0x-org/xpaywall/control-api/config"
	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

type errorResponse struct {
	Error string `json:"error"`
}

type Handler struct {
	cfg *config.ControlAPIConfig
	q   *postgres.Queries
	db  postgres.DBTX
}

func New(cfg *config.ControlAPIConfig, q *postgres.Queries, db postgres.DBTX) *Handler {
	return &Handler{cfg: cfg, q: q, db: db}
}
