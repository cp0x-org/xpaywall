package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli/v3"

	appconfig "github.com/cp0x-org/xpaywall/control-api/config"
)

const defaultMigrationsDir = "migrations"

func migrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "apply database migrations (goose up)",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dir", Value: defaultMigrationsDir, Usage: "migrations directory"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := appconfig.NewControlAPIConfig()
			if err != nil {
				return err
			}
			return runMigrations(ctx, cfg.DB_DSN, cmd.String("dir"))
		},
	}
}

func runMigrations(ctx context.Context, dsn, dir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, dir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Println("migrations applied")
	return nil
}
