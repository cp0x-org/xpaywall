#!/usr/bin/env bash

set -e

SUPERADMIN="torys"
PASSWORD="1cntgkth1"
EMAIL="torylandua@gmail.com"

echo "==> Applying migrations..."
go run cmd/control-api/main.go migrate

echo ""
echo "==> Installing payment methods..."

if ! go run cmd/control-api/main.go install payment-methods --superadmin "$SUPERADMIN"; then
    echo ""
    echo "Superadmin not found. Creating..."

    go run cmd/control-api/main.go install user \
        --username "$SUPERADMIN" \
        --password "$PASSWORD" \
        --role superadmin \
        --email "$EMAIL"

    echo ""
    echo "Retrying payment methods installation..."

    go run cmd/control-api/main.go install payment-methods --superadmin "$SUPERADMIN"
fi

echo ""
echo "==> Installing demo data..."
go run cmd/control-api/main.go install demo

echo ""
echo "✅ Done."