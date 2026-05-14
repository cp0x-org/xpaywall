-- Seed script: 75 request_logs rows over 7 days (2026-05-05 to 2026-05-12)
-- project_id: 4719a761-a5a1-4698-9d5c-002a9db5fc4c

DO $$
DECLARE
    p_id UUID := '4719a761-a5a1-4698-9d5c-002a9db5fc4c';
    routes UUID[] := ARRAY[
        'ff23a883-750e-4650-94f3-f861cca9dbd4'::UUID,
        '1356b75e-9099-4491-b3bb-a1eed2b0f907'::UUID,
        'd6afefb9-7453-4fb8-b5c4-9ccdf4a22ac5'::UUID,
        '4818322a-3fda-4da0-a1c3-66d403325016'::UUID,
        '80d45f13-0c77-4740-a562-6363120c0059'::UUID,
        '20dfe374-807e-4042-8761-5f3146f486ed'::UUID
    ];
    paths TEXT[] := ARRAY[
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
        '/free-multipoint/v3'
    ];
    methods TEXT[] := ARRAY['GET', 'GET', 'GET', 'POST', 'POST', 'PUT', 'DELETE'];
    statuses TEXT[] := ARRAY[
        'completed', 'completed', 'completed', 'completed',
        'failed', 'payment_required', 'payment_completed',
        'upstream_error', 'upstream_timeout', 'proxying'
    ];
    ips TEXT[] := ARRAY[
        '192.168.1.10', '10.0.0.5', '172.16.0.22', '203.0.113.45',
        '198.51.100.7', '185.199.108.1', '91.108.4.15', '66.249.64.1',
        '104.16.0.10', '151.101.0.81'
    ];
    agents TEXT[] := ARRAY[
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36',
        'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15',
        'curl/8.5.0',
        'python-httpx/0.27.0',
        'axios/1.6.8',
        'Go-http-client/1.1',
        'PostmanRuntime/7.37.0',
        'Mozilla/5.0 (X11; Linux x86_64; rv:125.0) Gecko/20100101 Firefox/125.0'
    ];

    st TEXT;
    pay_req BOOLEAN;
    pay_done BOOLEAN;
    req_at TIMESTAMP;
    done_at TIMESTAMP;
    amt_paid BIGINT;
    amt_usd NUMERIC(18,8);
    upstream_status INT;
    response_ms INT;
    final_code INT;
    err_type TEXT;
    err_msg TEXT;
    created_ts TIMESTAMP;
    rid UUID;
    route_idx INT;
    path_idx INT;
BEGIN
    FOR i IN 1..75 LOOP
        rid := gen_random_uuid();

        -- spread across 7 days with varied hours
        created_ts := TIMESTAMP '2026-05-01 00:00:00'
            + (random() * INTERVAL '7 days')
            + (random() * INTERVAL '23 hours')
            + (random() * INTERVAL '59 minutes');

        route_idx := 1 + (i % 6);
        path_idx  := 1 + (i % array_length(paths, 1));

        st := statuses[1 + (i % array_length(statuses, 1))];

        -- payment fields
        pay_req  := st IN ('payment_required', 'payment_completed', 'completed');
        pay_done := st IN ('payment_completed', 'completed') AND random() > 0.3;

        req_at   := CASE WHEN pay_req  THEN created_ts + (random() * INTERVAL '2 seconds') ELSE NULL END;
        done_at  := CASE WHEN pay_done THEN req_at    + (random() * INTERVAL '5 seconds') ELSE NULL END;

        amt_paid := CASE WHEN pay_done THEN (100000 + floor(random() * 900000))::BIGINT ELSE NULL END;
        amt_usd  := CASE WHEN pay_done THEN round((amt_paid / 1e6)::NUMERIC, 8) ELSE NULL END;

        -- upstream
        response_ms    := CASE
            WHEN st = 'upstream_timeout' THEN 10000 + floor(random() * 5000)::INT
            WHEN st IN ('completed', 'proxying') THEN 50 + floor(random() * 500)::INT
            ELSE NULL
        END;
        upstream_status := CASE
            WHEN st = 'completed'      THEN (ARRAY[200,200,200,201,204])[1 + floor(random()*5)::INT]
            WHEN st = 'upstream_error' THEN (ARRAY[500,502,503,504])[1 + floor(random()*4)::INT]
            WHEN st = 'proxying'       THEN 200
            ELSE NULL
        END;
        final_code := CASE
            WHEN st = 'completed'        THEN upstream_status
            WHEN st = 'payment_required' THEN 402
            WHEN st = 'payment_completed'THEN 200
            WHEN st = 'failed'           THEN 400
            WHEN st = 'upstream_error'   THEN upstream_status
            WHEN st = 'upstream_timeout' THEN 504
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
            amount_paid, amount_usd,
            upstream_url, upstream_status_code, upstream_response_time_ms,
            final_status_code, error_type, error_message,
            created_at, updated_at
        ) VALUES (
            rid, p_id, routes[route_idx], 'req-' || replace(rid::TEXT, '-', ''),
            methods[1 + (i % array_length(methods, 1))],
            paths[path_idx],
            ips[1 + (i % array_length(ips, 1))],
            agents[1 + (i % array_length(agents, 1))],
            st, pay_req, req_at,
            pay_done, done_at,
            amt_paid, amt_usd,
            'https://upstream.example.com' || paths[path_idx],
            upstream_status, response_ms,
            final_code, err_type, err_msg,
            created_ts, created_ts
        );
    END LOOP;

    RAISE NOTICE 'Inserted 75 request_logs rows';
END $$;
