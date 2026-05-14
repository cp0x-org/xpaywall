package app

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	appconfig "github.com/cp0x-org/xpaywall/control-api/config"
	internalhttp "github.com/cp0x-org/xpaywall/control-api/internal/http"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers"
	authhandler "github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/auth"
	"github.com/cp0x-org/xpaywall/control-api/internal/http/handlers/gateway"
	postgres "github.com/cp0x-org/xpaywall/control-api/internal/storage/postgres/generated"
)

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:     "control-api",
		Usage:    "run xpaywall control api",
		Commands: []*cli.Command{installCommand()},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "env-file",
				Usage: "path to .env file",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			return ctx, appconfig.LoadEnv(cmd.String("env-file"))
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			_ = cmd

			cfg, err := appconfig.NewControlAPIConfig()
			if err != nil {
				return err
			}

			return run(ctx, cfg)
		},
	}
}

func Run(ctx context.Context, args []string) error {
	return NewCommand().Run(ctx, args)
}

func run(ctx context.Context, cfg *appconfig.ControlAPIConfig) error {
	gin.SetMode(cfg.Mode)

	pool, err := pgxpool.New(ctx, cfg.DB_DSN)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	q := postgres.New(pool)
	h := handlers.New(cfg, q, pool)
	ah := authhandler.New(q, cfg.JWTSecret, cfg.SuperadminUsername, cfg.SuperadminPassword)
	gw := gateway.New(q)

	router := internalhttp.SetupRouter(h, ah, gw, cfg.InternalAPIKey, cfg.JWTSecret)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("control-api listening on %s", addr)

	return router.Run(addr)
}
