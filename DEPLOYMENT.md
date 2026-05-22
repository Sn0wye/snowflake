# Deployment Configuration

All services use committed config files as defaults with environment variable overrides. No separate prod configs — each service reads one config file everywhere (local.yml for Go, application.properties for Java, appsettings.json for .NET).

## Services

- **helium** (Go): `config/local.yml` + env overrides (dots→underscores)
- **gold** (Go): `config/local.yml` + env overrides (dots→underscores)
- **carbon** (Java): `application.properties` + Spring `SPRING_*` env vars
- **oxygen** (.NET): `appsettings.json` + double-underscore env vars

## Config Implementation

**helium & gold** (Go/Viper): Automatically binds env vars with underscore replacement. Specific bindings: `DATABASE_URL`, `RABBITMQ_URL`, `JWT_SECRET` map to YAML paths. Override via `APP_CONF=/path/to/config.yml`.

**carbon** (Spring Boot): Natively binds `SPRING_*` env vars. Exception: `SPRING_DATASOURCE_URL` requires JDBC format (`jdbc:postgresql://...`).

**oxygen** (.NET): Uses double-underscore notation in env vars (e.g., `ConnectionStrings__DefaultConnection` → `ConnectionStrings:DefaultConnection`).

## Environment Variables

**Quick reference:** Copy `.env.example` for local testing.

Key naming conventions:
- **helium/gold:** `{SERVICE}_*` prefix (e.g., `HELIUM_DATABASE_URL`, `GOLD_JWT_SECRET`)
- **carbon:** `SPRING_*` prefix (e.g., `SPRING_DATASOURCE_URL`, `LOGGING_LEVEL_ROOT`)
- **oxygen:** CamelCase with double-underscores (e.g., `ConnectionStrings__DefaultConnection`, `Logging__LogLevel__Default`)

## Docker Compose (Local Dev Only)

`docker-compose.yml` is for local development only — not used in production. All `environment:` blocks use `${SERVICE_VAR:-default}` so unset vars fall back to service config file defaults. Includes nginx as a local reverse proxy.

## Local Development

**Option A: No docker (fastest):**
```bash
# Services read from local config files (no env vars needed)
docker compose up postgres rabbitmq  # infra only
go run ./helium/src/cmd/server.go    # helium
```

**Option B: Full docker-compose (includes nginx proxy on :80):**
```bash
# Copy .env.example → .env, update as needed
docker compose up
```

## Production Deployment

Each service is deployed **independently** via CI/CD:

1. Push to `main` triggers the service's GitHub Actions workflow (`.github/workflows/publish-<service>.yml`)
2. Workflow builds the Docker image and pushes to `ghcr.io/<repo>/<service>:prod`
3. Workflow calls the Coolify webhook (`COOLIFY_<SERVICE>_WEBHOOK`) which redeploys that service
4. Coolify pulls the new image and restarts the container with its configured env vars

Each service is configured separately in Coolify with its own env vars. There is no shared docker-compose in production — nginx is not used in prod.

## Configuration Priority (lowest to highest)

1. **Config file defaults** (local.yml, application.properties, appsettings.json)
2. **Environment variables** (override file defaults)
3. **Explicit env var bindings** (e.g., `DATABASE_URL` → `db.connectionString`)

## Format Notes

Each service uses its native driver's URL format — intentional, not a bug:
- **Go (GORM):** `postgres://user:pass@host:5432/db?sslmode=disable`
- **Java (JDBC):** `jdbc:postgresql://host:5432/db`
- **.NET (Npgsql):** `Host=host;Database=db;Username=user;Password=pass`

When setting Coolify env vars, use the correct format for each service.
