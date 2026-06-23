package app

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
				Usage: "seed demo data (admin user, sample project, routes, request logs)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "admin-username", Value: "admin", Usage: "admin login username"},
					&cli.StringFlag{Name: "admin-password", Value: "admin", Usage: "admin plaintext password (will be hashed)"},
					&cli.BoolFlag{Name: "skip-logs", Usage: "skip request_logs seed"},
					&cli.IntFlag{Name: "log-days", Value: 7, Usage: "spread log entries over this many past days"},
					&cli.IntFlag{Name: "log-count", Value: 75, Usage: "number of request_logs rows to insert"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := appconfig.NewControlAPIConfig()
					if err != nil {
						return err
					}
					if err := runDemoSeed(ctx, cfg.DB_DSN, cmd.String("admin-username"), cmd.String("admin-password")); err != nil {
						return err
					}
					if cmd.Bool("skip-logs") {
						return nil
					}
					return runSeedLogs(ctx, cfg.DB_DSN, "default", demoLogPaths, int(cmd.Int("log-days")), int(cmd.Int("log-count")))
				},
			},
			{
				Name:  "demo-mpp",
				Usage: "seed MPP/Tempo charge demo data (admin user, MPP project, routes, request logs)",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "admin-username", Value: "admin", Usage: "admin login username"},
					&cli.StringFlag{Name: "admin-password", Value: "admin", Usage: "admin plaintext password (will be hashed)"},
					&cli.BoolFlag{Name: "skip-logs", Usage: "skip request_logs seed"},
					&cli.IntFlag{Name: "log-days", Value: 7, Usage: "spread log entries over this many past days"},
					&cli.IntFlag{Name: "log-count", Value: 75, Usage: "number of request_logs rows to insert"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := appconfig.NewControlAPIConfig()
					if err != nil {
						return err
					}
					if err := runDemoMPPSeed(ctx, cfg.DB_DSN, cmd.String("admin-username"), cmd.String("admin-password")); err != nil {
						return err
					}
					if cmd.Bool("skip-logs") {
						return nil
					}
					return runSeedLogs(ctx, cfg.DB_DSN, mppProjectSlug, demoMPPLogPaths, int(cmd.Int("log-days")), int(cmd.Int("log-count")))
				},
			},
			{
				Name:  "user",
				Usage: "create a user with a bcrypt-hashed password",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "username", Required: true, Usage: "login username"},
					&cli.StringFlag{Name: "password", Required: true, Usage: "plaintext password (will be hashed)"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := appconfig.NewControlAPIConfig()
					if err != nil {
						return err
					}
					return runCreateUser(ctx, cfg.DB_DSN, cmd.String("username"), cmd.String("password"))
				},
			},
		},
	}
}

const demoSQL = `
DO $$
DECLARE
    v_admin_id       UUID := '%s';
    v_project_id     UUID;
    v_method_id      UUID;
    v_asset_id       UUID;
    v_facilitator_id UUID;
BEGIN
    INSERT INTO projects (id, owner_user_id, name, slug)
    VALUES (gen_random_uuid(), v_admin_id, 'Default Project', 'default')
    ON CONFLICT (slug) WHERE archived_at IS NULL DO NOTHING
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
    ) ON CONFLICT (project_id, payment_method_id, asset_id, scheme) DO NOTHING;

    IF NOT EXISTS (SELECT 1 FROM routes WHERE project_id = v_project_id) THEN
        INSERT INTO routes (
            id, project_id, name, path_pattern, price_usd, description, free
        ) VALUES
            (gen_random_uuid(), v_project_id, 'health',             '/health',            '0.001', 'Health check endpoint',                              FALSE),
            (gen_random_uuid(), v_project_id, 'metered-api',        '/api/metered/*',     '0.10',  'Metered API billed by actual usage',                  FALSE),
            (gen_random_uuid(), v_project_id, 'weather',            '/weather',           '0.10',  'Get weather data',                                    FALSE),
            (gen_random_uuid(), v_project_id, 'free-endpoint',      '/free-endpoint',     '',      'Free endpoint, no payment required',                  TRUE),
            (gen_random_uuid(), v_project_id, 'free-multipoint',    '/free-multipoint',   '',      'Free multipoint root, no payment required',           TRUE),
            (gen_random_uuid(), v_project_id, 'free-multipoint-sub','/free-multipoint/*', '',      'Free multipoint with subpath, no payment required',   TRUE);
    END IF;
END;
$$;
`

// mppProjectSlug is the slug of the project created by `install demo-mpp`.
const mppProjectSlug = "mpp-demo"

// Log path-sets seeded by runSeedLogs. Each demo seeds logs against its own
// (disjoint) upstream endpoints so the two projects can be tested independently.
const (
	demoLogPaths = `
        '/health',
        '/api/metered/data',
        '/api/metered/query',
        '/api/metered/report',
        '/weather',
        '/weather?city=London',
        '/weather?city=New+York',
        '/free-endpoint',
        '/free-multipoint',
        '/free-multipoint/v2',
        '/free-multipoint/v3'`

	demoMPPLogPaths = `
        '/time',
        '/api/usage/data',
        '/api/usage/query',
        '/api/usage/report',
        '/quote',
        '/quote?symbol=BTC',
        '/quote?symbol=ETH',
        '/ping',
        '/echo',
        '/echo/v2',
        '/echo/v3'`
)

// demoMPPSQL seeds a self-contained MPP / Tempo "charge" demo: a dedicated
// project, upstream settings, an MPP payment method (pathUSD stablecoin on
// Tempo) with method/scheme set, the project↔method link carrying
// rpc_url/secret_key in its config JSONB and no facilitator, plus paid + free
// routes. Prices are stored as USD strings; xgateway converts them to raw base
// units (decimals=6) and then to the human charge amount.
const demoMPPSQL = `
DO $$
DECLARE
    v_admin_id   UUID := '%s';
    v_project_id UUID;
    v_method_id  UUID;
    v_asset_id   UUID;
BEGIN
    INSERT INTO projects (id, owner_user_id, name, slug)
    VALUES (gen_random_uuid(), v_admin_id, 'Tempo MPP Project', 'mpp-demo')
    ON CONFLICT (slug) WHERE archived_at IS NULL DO NOTHING
    RETURNING id INTO v_project_id;

    IF v_project_id IS NULL THEN
        SELECT id INTO v_project_id FROM projects WHERE slug = 'mpp-demo';
    END IF;

    INSERT INTO project_routes_settings (
        id, project_id, base_url, auth_header_name, auth_header_value, allow_unmatched
    ) VALUES (
        gen_random_uuid(), v_project_id, 'http://localhost:4021',
        'Authorization', 'Bearer YOUR_UPSTREAM_ACCESS_TOKEN', FALSE
    ) ON CONFLICT (project_id) DO NOTHING;

    INSERT INTO payment_methods (id, code, protocol, name, caip2_chain_id, method, scheme, enabled)
    VALUES (gen_random_uuid(), 'mpp_tempo_charge', 'mpp', 'Tempo Charge', NULL, 'tempo', 'charge', TRUE)
    ON CONFLICT (code) DO UPDATE SET enabled = TRUE, method = EXCLUDED.method, scheme = EXCLUDED.scheme
    RETURNING id INTO v_method_id;

    -- pathUSD: the Tempo stablecoin used for charges.
    INSERT INTO payment_method_assets (id, payment_method_id, symbol, contract_address, decimals)
    VALUES (gen_random_uuid(), v_method_id, 'pathUSD', '0x20c0000000000000000000000000000000000000', 6)
    ON CONFLICT (payment_method_id, symbol) DO UPDATE SET
        contract_address = EXCLUDED.contract_address,
        decimals         = EXCLUDED.decimals
    RETURNING id INTO v_asset_id;

    -- MPP has no facilitator: facilitator_id stays NULL; rpc_url/secret_key live
    -- in the config JSONB consumed by xgateway's HTTP provider.
    INSERT INTO project_payment_methods (
        id, project_id, payment_method_id, asset_id, scheme, payout_address, config, enabled
    ) VALUES (
        gen_random_uuid(), v_project_id, v_method_id, v_asset_id,
        'charge', '0xb7Ac41661288791225d66643a7d952e551FC7868',
        '{"method":"tempo","rpc_url":"https://rpc.moderato.tempo.xyz","secret_key":"xgateway-dev-secret"}'::jsonb,
        TRUE
    ) ON CONFLICT (project_id, payment_method_id, asset_id, scheme) DO NOTHING;

    IF NOT EXISTS (SELECT 1 FROM routes WHERE project_id = v_project_id) THEN
        INSERT INTO routes (
            id, project_id, name, path_pattern, price_usd, description, free
        ) VALUES
            (gen_random_uuid(), v_project_id, 'time',      '/time',         '0.001', 'Server time endpoint (Tempo charge)',               FALSE),
            (gen_random_uuid(), v_project_id, 'usage-api', '/api/usage/*',  '0.10',  'Usage-metered API billed via Tempo charge',         FALSE),
            (gen_random_uuid(), v_project_id, 'quote',     '/quote',        '0.10',  'Get a price quote, paid via Tempo charge',          FALSE),
            (gen_random_uuid(), v_project_id, 'ping',      '/ping',         '',      'Free ping endpoint, no payment required',           TRUE),
            (gen_random_uuid(), v_project_id, 'echo',      '/echo',         '',      'Free echo root, no payment required',               TRUE),
            (gen_random_uuid(), v_project_id, 'echo-sub',  '/echo/*',       '',      'Free echo with subpath, no payment required',       TRUE);
    END IF;
END;
$$;
`

const seedLogsSQL = `
DO $$
DECLARE
    p_id      UUID;
    route_ids UUID[];
    n_routes  INT;
    n_logs    INT := %d;
    days_back INT := %d;
    paths    TEXT[] := ARRAY[%s];
    methods  TEXT[] := ARRAY['GET', 'GET', 'GET', 'POST', 'POST', 'PUT', 'DELETE'];
    statuses TEXT[] := ARRAY[
        'completed', 'completed', 'completed', 'completed',
        'failed', 'payment_required', 'payment_completed',
        'upstream_error', 'upstream_timeout', 'proxying'
    ];
    ips      TEXT[] := ARRAY[
        '192.168.1.10', '10.0.0.5', '172.16.0.22', '203.0.113.45',
        '198.51.100.7', '185.199.108.1', '91.108.4.15', '66.249.64.1',
        '104.16.0.10', '151.101.0.81'
    ];
    agents   TEXT[] := ARRAY[
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36',
        'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15',
        'curl/8.5.0',
        'python-httpx/0.27.0',
        'axios/1.6.8',
        'Go-http-client/1.1',
        'PostmanRuntime/7.37.0',
        'Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0'
    ];

    st              TEXT;
    pay_req         BOOLEAN;
    pay_done        BOOLEAN;
    req_at          TIMESTAMP;
    done_at         TIMESTAMP;
    amt_usd         NUMERIC(18,8);
    upstream_status INT;
    response_ms     INT;
    final_code      INT;
    err_type        TEXT;
    err_msg         TEXT;
    created_ts      TIMESTAMP;
    rid             UUID;
    path_idx        INT;
BEGIN
    SELECT id INTO p_id FROM projects WHERE slug = '%s';
    IF p_id IS NULL THEN
        RAISE NOTICE 'no project found for slug, skipping request_logs seed';
        RETURN;
    END IF;

    SELECT array_agg(id) INTO route_ids FROM routes WHERE project_id = p_id;
    n_routes := COALESCE(array_length(route_ids, 1), 0);
    IF n_routes = 0 THEN
        RAISE NOTICE 'no routes in default project, skipping request_logs seed';
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM request_logs WHERE project_id = p_id LIMIT 1) THEN
        RAISE NOTICE 'request_logs already populated for default project, skipping seed';
        RETURN;
    END IF;

    FOR i IN 1..n_logs LOOP
        rid := gen_random_uuid();

        created_ts := NOW() - (random() * days_back) * INTERVAL '1 day';

        path_idx := 1 + (i %% array_length(paths, 1));
        st       := statuses[1 + (i %% array_length(statuses, 1))];

        pay_req  := st IN ('payment_required', 'payment_completed', 'completed');
        pay_done := st IN ('payment_completed', 'completed') AND random() > 0.3;

        req_at  := CASE WHEN pay_req  THEN created_ts + (random() * INTERVAL '2 seconds') ELSE NULL END;
        done_at := CASE WHEN pay_done THEN req_at    + (random() * INTERVAL '5 seconds') ELSE NULL END;

        amt_usd := CASE WHEN pay_done THEN round(((100000 + floor(random() * 900000))::NUMERIC / 1e6), 8) ELSE NULL END;

        response_ms := CASE
            WHEN st = 'upstream_timeout'          THEN 10000 + floor(random() * 5000)::INT
            WHEN st IN ('completed', 'proxying')  THEN 50 + floor(random() * 500)::INT
            ELSE NULL
        END;
        upstream_status := CASE
            WHEN st = 'completed'      THEN (ARRAY[200,200,200,201,204])[1 + floor(random()*5)::INT]
            WHEN st = 'upstream_error' THEN (ARRAY[500,502,503,504])[1 + floor(random()*4)::INT]
            WHEN st = 'proxying'       THEN 200
            ELSE NULL
        END;
        final_code := CASE
            WHEN st = 'completed'         THEN upstream_status
            WHEN st = 'payment_required'  THEN 402
            WHEN st = 'payment_completed' THEN 200
            WHEN st = 'failed'            THEN 400
            WHEN st = 'upstream_error'    THEN upstream_status
            WHEN st = 'upstream_timeout'  THEN 504
            ELSE 200
        END;

        err_type := CASE
            WHEN st = 'failed'           THEN 'bad_request'
            WHEN st = 'upstream_error'   THEN 'upstream_error'
            WHEN st = 'upstream_timeout' THEN 'upstream_timeout'
            ELSE NULL
        END;
        err_msg := CASE
            WHEN st = 'failed'           THEN 'invalid request parameters'
            WHEN st = 'upstream_error'   THEN 'upstream returned ' || upstream_status::TEXT
            WHEN st = 'upstream_timeout' THEN 'upstream timed out after ' || response_ms::TEXT || 'ms'
            ELSE NULL
        END;

        INSERT INTO request_logs (
            id, project_id, outbound_route_id, request_id,
            method, path, client_ip, user_agent,
            status, payment_required, payment_requested_at,
            payment_completed, payment_completed_at,
            amount_usd,
            upstream_url, upstream_status_code, upstream_response_time_ms,
            final_status_code, error_type, error_message,
            created_at, updated_at
        ) VALUES (
            rid, p_id, route_ids[1 + (i %% n_routes)], 'req-' || replace(rid::TEXT, '-', ''),
            methods[1 + (i %% array_length(methods, 1))],
            paths[path_idx],
            ips[1 + (i %% array_length(ips, 1))],
            agents[1 + (i %% array_length(agents, 1))],
            st, pay_req, req_at,
            pay_done, done_at,
            amt_usd,
            'https://upstream.example.com' || paths[path_idx],
            upstream_status, response_ms,
            final_code, err_type, err_msg,
            created_ts, created_ts
        );
    END LOOP;
END $$;
`

func runCreateUser(ctx context.Context, dsn, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	const stmt = `
INSERT INTO users (id, username, password_hash)
VALUES (gen_random_uuid(), $1, crypt($2, gen_salt('bf', 10)))
ON CONFLICT (username) DO NOTHING
RETURNING id`

	var id uuid.UUID
	err = pool.QueryRow(ctx, stmt, username, password).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("user with username %q already exists", username)
	}
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	log.Printf("user created: %s (id: %s)", username, id)
	return nil
}

func runDemoSeed(ctx context.Context, dsn, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("admin-username and admin-password are required")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	adminID, err := ensureAdminUser(ctx, pool, username, password)
	if err != nil {
		return err
	}

	if _, err = pool.Exec(ctx, fmt.Sprintf(demoSQL, adminID)); err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}

	log.Println("demo data installed")
	log.Printf("  user:    %s / %s", username, password)
	log.Println("  project: Default Project (slug: default)")
	return nil
}

func runDemoMPPSeed(ctx context.Context, dsn, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("admin-username and admin-password are required")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	adminID, err := ensureAdminUser(ctx, pool, username, password)
	if err != nil {
		return err
	}

	if _, err = pool.Exec(ctx, fmt.Sprintf(demoMPPSQL, adminID)); err != nil {
		return fmt.Errorf("seed mpp demo data: %w", err)
	}

	log.Println("mpp demo data installed")
	log.Printf("  user:    %s / %s", username, password)
	log.Printf("  project: Tempo MPP Project (slug: %s)", mppProjectSlug)
	log.Println("  payment: MPP / Tempo charge (pathUSD stablecoin)")
	return nil
}

// ensureAdminUser inserts the admin user (or fetches its id when it already
// exists) and returns the user id used as the demo project owner.
func ensureAdminUser(ctx context.Context, pool *pgxpool.Pool, username, password string) (uuid.UUID, error) {
	const userStmt = `
INSERT INTO users (id, username, password_hash)
VALUES (gen_random_uuid(), $1, crypt($2, gen_salt('bf', 10)))
ON CONFLICT (username) DO NOTHING
RETURNING id`

	var adminID uuid.UUID
	err := pool.QueryRow(ctx, userStmt, username, password).Scan(&adminID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = pool.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, username).Scan(&adminID)
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create or fetch admin user: %w", err)
	}
	return adminID, nil
}

func runSeedLogs(ctx context.Context, dsn, slug, paths string, days, count int) error {
	if days <= 0 {
		return fmt.Errorf("log-days must be > 0")
	}
	if count <= 0 {
		return fmt.Errorf("log-count must be > 0")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	if _, err = pool.Exec(ctx, fmt.Sprintf(seedLogsSQL, count, days, paths, slug)); err != nil {
		return fmt.Errorf("seed request_logs: %w", err)
	}

	log.Printf("request_logs seeded for project %q: %d rows over last %d days", slug, count, days)
	return nil
}
