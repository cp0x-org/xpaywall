# xpaywall control-api

Control API service for xpaywall. This project owns the database schema, migrations, and GORM models.

## Run

```powershell
$env:CONTROL_DB_DSN = "postgres://user:pass@localhost:5432/xpaywall?sslmode=disable"
go run ./cmd/control-api
```

You can also load env vars from a file:

```powershell
go run ./cmd/control-api --env-file .env
```

## Migrations

```powershell
go tool goose -dir migrations postgres "$env:CONTROL_DB_DSN" status
go tool goose -dir migrations postgres "$env:CONTROL_DB_DSN" up
go tool goose -dir migrations postgres "$env:CONTROL_DB_DSN" down
```

## Test

```powershell
go test ./...
```
