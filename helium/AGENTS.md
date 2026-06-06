# AGENTS.md — Helium

Go/Fiber auth service. HTTP on 8080, gRPC on 50050.

## Quick Commands

```bash
# Run server (hot reload)
air

# Build all binaries
go build ./...

# Run migrations
make migrate-up
make migrate-down

# Create new migration
make migrate-create NAME=add_foo_column

# Regenerate proto + OpenAPI
make generate
make docs
```

## Database Migrations

Migrations use **golang-migrate** (v4) with `embed.FS` — SQL files are compiled into the binary. No external files or path dependencies.

### Architecture

| File | Role |
|------|------|
| `src/cmd/migrate/main.go` | Migration entry point (separate binary from server) |
| `src/migration/migration.go` | Runner: Up/Down via golang-migrate + iofs |
| `src/migration/migrations/` | SQL migration files (embed.FS) |

### Binary

Two binaries are built:

```
src/cmd/server/main.go   → main   (HTTP + gRPC server)
src/cmd/migrate/main.go  → migrate (migrate up/down)
```

Server **does not** run AutoMigrate. Migrations must be run explicitly before server start.

### Makefile Targets

```bash
# Apply pending migrations
# Requires DATABASE_URL env var or config/local.yml
make migrate-up

# Rollback last migration (one step)
make migrate-down

# Create new migration pair
# Generates 000002_<name>.{up,down}.sql in src/migration/migrations/
make migrate-create NAME=add_user_status
```

`DATABASE_URL` defaults to `postgres://snowflake:snowflake@127.0.0.1:5432/helium?sslmode=disable`. Override:

```bash
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require make migrate-up
```

### Writing Migrations

1. Create a new migration pair:
   ```bash
   make migrate-create NAME=add_recovery_email_to_users
   ```

2. Edit the generated `.up.sql` and `.down.sql` files in `src/migration/migrations/`.

3. Update the corresponding GORM model (optional but recommended for consistency).

4. Rebuild and test:
   ```bash
   make migrate-up
   make migrate-down
   make migrate-up
   ```

### Production

Both binaries are built in CI (`publish-helium.yml`) and copied into the Docker image:

```bash
# In production (Coolify):
./migrate up
./main
```

Migrations are embedded — no volume mounts or file paths needed.

### Rollback Safety

`migrate down` uses `m.Steps(-1)` — rolls back exactly **one** migration, not all of them. Safe for production.

## Configuration

- Config: `config/local.yml` (default). Override with `APP_CONF`.
- Env vars: `DATABASE_URL`, `RABBITMQ_URL`, `JWT_SECRET`, `JWT_ISSUER`, `REFRESH_TOKEN_SECRET`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`.
- Dots → underscores for all Viper keys (`HTTP_PORT` → `http.port`).

## Docker

```bash
# Build
docker build -t helium .

# Run (needs DATABASE_URL)
docker run -e DATABASE_URL=postgres://... helium
```

CI Dockerfile (`Dockerfile.ci`) is separate from local dev Dockerfile — CI pre-builds binaries with Go, then copies into a minimal Alpine image.

## Key Files

| Path | Purpose |
|------|---------|
| `src/cmd/server/main.go` | Server entry point |
| `src/cmd/migrate/main.go` | Migration entry point |
| `src/migration/migration.go` | Migration runner (embed + iofs) |
| `src/migration/migrations/` | SQL migration files |
| `src/models/models.go` | GORM model registry |
| `src/db/db.go` | GORM DB singleton |
| `pkg/config/config.go` | Viper config loader |
| `config/local.yml` | Default config |
| `Makefile` | Proto gen, OpenAPI gen, migrations |
| `.air.toml` | Air hot reload config |
