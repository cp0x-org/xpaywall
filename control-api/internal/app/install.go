package app

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	appconfig "github.com/cp0x-org/xpaywall/control-api/config"
)

func installCommand() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "installation utilities",
		Commands: []*cli.Command{
			{
				Name:  "demo",
				Usage: "seed demo data (admin user + sample project and routes)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := appconfig.NewControlAPIConfig()
					if err != nil {
						return err
					}
					return runDemoSeed(ctx, cfg.DB_DSN)
				},
			},
		},
	}
}

const demoSQL = `
DO $$
DECLARE
    v_admin_id       UUID;
    v_project_id     UUID;
    v_method_id      UUID;
    v_asset_id       UUID;
    v_facilitator_id UUID;
BEGIN
    INSERT INTO users (id, username, password_hash)
    VALUES (gen_random_uuid(), 'admin', crypt('admin', gen_salt('bf', 10)))
    ON CONFLICT (username) DO NOTHING
    RETURNING id INTO v_admin_id;

    IF v_admin_id IS NULL THEN
        SELECT id INTO v_admin_id FROM users WHERE username = 'admin';
    END IF;

    INSERT INTO projects (id, owner_user_id, name, slug)
    VALUES (gen_random_uuid(), v_admin_id, 'Default Project', 'default')
    ON CONFLICT (slug) DO NOTHING
    RETURNING id INTO v_project_id;

    IF v_project_id IS NULL THEN
        SELECT id INTO v_project_id FROM projects WHERE slug = 'default';
    END IF;

    INSERT INTO project_routes_settings (
        id, project_id, base_url, auth_header_name, auth_header_value, allow_unmatched
    ) VALUES (
        gen_random_uuid(), v_project_id, 'http://localhost:4021',
        'Authorization', 'Bearer YOUR_UPSTREAM_ACCESS_TOKEN', FALSE
    ) ON CONFLICT (project_id) DO NOTHING;

    INSERT INTO payment_methods (id, code, protocol, name, caip2_chain_id, enabled)
    VALUES (gen_random_uuid(), 'x402_base_sepolia', 'x402', 'Base Sepolia', 'eip155:84532', TRUE)
    ON CONFLICT (code) DO UPDATE SET enabled = TRUE
    RETURNING id INTO v_method_id;

    INSERT INTO payment_method_assets (id, payment_method_id, symbol, contract_address, decimals)
    VALUES (gen_random_uuid(), v_method_id, 'USDC', '0x036CbD53842c5426634e7929541eC2318f3dCF7e', 6)
    ON CONFLICT (payment_method_id, symbol) DO UPDATE SET decimals = EXCLUDED.decimals
    RETURNING id INTO v_asset_id;

    SELECT id INTO v_facilitator_id FROM facilitators WHERE name = 'x402 Coinbase' LIMIT 1;

    IF v_facilitator_id IS NULL THEN
        INSERT INTO facilitators (id, name, url, enabled)
        VALUES (gen_random_uuid(), 'x402 Coinbase', 'https://x402.org/facilitator', TRUE)
        RETURNING id INTO v_facilitator_id;
    END IF;

    INSERT INTO project_payment_methods (
        id, project_id, payment_method_id, asset_id, scheme, facilitator_id, payout_address, enabled
    ) VALUES (
        gen_random_uuid(), v_project_id, v_method_id, v_asset_id,
        'exact', v_facilitator_id, '0xEb6ae6fA22D307Eae06BE0862087FdFFdD25Bab4', TRUE
    );

    INSERT INTO routes (
        id, project_id, name, path_pattern, price_amount, price_usd, description, free
    ) VALUES
        (gen_random_uuid(), v_project_id, 'health',             '/health',            0, '$0.001', 'Health check endpoint',                              FALSE),
        (gen_random_uuid(), v_project_id, 'metered-api',        '/api/metered/*',     0, '$0.10',  'Metered API billed by actual usage',                  FALSE),
        (gen_random_uuid(), v_project_id, 'weather',            '/weather',           0, '$0.10',  'Get weather data',                                    FALSE),
        (gen_random_uuid(), v_project_id, 'free-endpoint',      '/free-endpoint',     0, '',       'Free endpoint, no payment required',                  TRUE),
        (gen_random_uuid(), v_project_id, 'free-multipoint',    '/free-multipoint',   0, '',       'Free multipoint root, no payment required',           TRUE),
        (gen_random_uuid(), v_project_id, 'free-multipoint-sub','/free-multipoint/*', 0, '',       'Free multipoint with subpath, no payment required',   TRUE);
END;
$$;
`

func runDemoSeed(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	if _, err = pool.Exec(ctx, demoSQL); err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}

	log.Println("demo data installed")
	log.Println("  user:    admin / admin")
	log.Println("  project: Default Project (slug: default)")
	return nil
}
