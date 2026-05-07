# Deployment Configuration Plan

## Overview

All microservices use **committed dev/local config files as defaults** and support **environment variable overrides for production**. In Coolify, you manage all overrides through the project environment variables UI (no `.env` files needed).

---

## Architecture

**Local/Dev (default):**
- helium: reads `config/local.yml`
- gold: reads `config/local.yml`
- carbon: reads `application-dev.properties`
- oxygen: reads `appsettings.json`

**Production (Coolify):**
- helium: reads `config/prod.yml`, overridden by Coolify-set env vars
- gold: reads `config/prod.yml`, overridden by Coolify-set env vars
- carbon: reads `application-prod.properties`, overridden by Spring Boot env var conventions
- oxygen: reads `appsettings.json`, overridden by .NET configuration conventions

---

## Code Changes Required

### Phase 1: helium & gold (`pkg/config/config.go`)

Add env var override support to both services. In the `getConfig(path string)` function, after reading the YAML file:

```go
// Enable environment variable overrides
conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
conf.AutomaticEnv()

// Explicit conventional bindings — guarded against empty string
if v := os.Getenv("DATABASE_URL"); v != "" {
    conf.Set("db.connectionString", v)
}
if v := os.Getenv("RABBITMQ_URL"); v != "" {
    conf.Set("messaging.connectionString", v)
}
if v := os.Getenv("JWT_SECRET"); v != "" {
    conf.Set("security.jwt_secret", v)
}
```

**What this enables:**
- `DATABASE_URL`, `RABBITMQ_URL`, `JWT_SECRET` override the YAML defaults
- All other YAML keys can be overridden via uppercased env vars with dots replaced by underscores
  - Example: `HTTP_PORT` → `http.port`, `LOG_LOG_LEVEL` → `log.log_level`

**carbon & oxygen:**
- No code changes needed — Spring Boot and .NET natively support env var overrides

---

## Docker Compose Configuration

### Add `environment:` blocks to each service

These blocks map Coolify-level service-prefixed variables to container-level env vars.

**helium:**
```yaml
environment:
  - DATABASE_URL=${HELIUM_DATABASE_URL:-}
  - RABBITMQ_URL=${HELIUM_RABBITMQ_URL:-}
  - JWT_SECRET=${HELIUM_JWT_SECRET:-}
  - JWT_ISSUER=${HELIUM_JWT_ISSUER:-}
  - HTTP_PORT=${HELIUM_HTTP_PORT:-8080}
  - GRPC_PORT=${HELIUM_GRPC_PORT:-50050}
  - LOG_LOG_LEVEL=${HELIUM_LOG_LEVEL:-info}
```

**gold:**
```yaml
environment:
  - DATABASE_URL=${GOLD_DATABASE_URL:-}
  - RABBITMQ_URL=${GOLD_RABBITMQ_URL:-}
  - JWT_SECRET=${GOLD_JWT_SECRET:-}
  - JWT_ISSUER=${GOLD_JWT_ISSUER:-}
  - HTTP_PORT=${GOLD_HTTP_PORT:-8083}
  - GRPC_PORT=${GOLD_GRPC_PORT:-50053}
  - LOG_LOG_LEVEL=${GOLD_LOG_LEVEL:-info}
```

**carbon** (Spring Boot conventions, uses JDBC URL format):
```yaml
environment:
  - SPRING_DATASOURCE_URL=${CARBON_DATABASE_URL:-}
  - SPRING_RABBITMQ_ADDRESSES=${CARBON_RABBITMQ_URL:-}
  - SPRING_GRPC_HOST=${CARBON_GRPC_HOST:-snowflake-helium}
  - SERVER_PORT=${CARBON_HTTP_PORT:-8081}
  - LOGGING_LEVEL_ROOT=${CARBON_LOG_LEVEL:-INFO}
```

**oxygen** (.NET conventions, uses double-underscore notation):
```yaml
environment:
  - ConnectionStrings__DefaultConnection=${OXYGEN_DATABASE_URL:-}
  - ConnectionStrings__UsersGrpc=${OXYGEN_GRPC_URL:-http://snowflake-helium:50050}
  - Security__Jwt__SecretKey=${OXYGEN_JWT_SECRET:-}
  - ASPNETCORE_HTTP_PORTS=${OXYGEN_HTTP_PORT:-8082}
  - Logging__LogLevel__Default=${OXYGEN_LOG_LEVEL:-Information}
```

---

## Coolify Environment Variables

Set these in Coolify's Docker Compose project via the **Environment Variables** UI.

### Helium

| Variable | Format | Example |
|----------|--------|---------|
| `HELIUM_DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | `postgres://snowflake:password@db.example.com:5432/helium?sslmode=require` |
| `HELIUM_RABBITMQ_URL` | `amqp://user:pass@host:5672` | `amqp://snowflake:password@mq.example.com:5672` |
| `HELIUM_JWT_SECRET` | any string (min 32 chars recommended) | `your-secret-key-here` |
| `HELIUM_JWT_ISSUER` | any string | `snowflake` |
| `HELIUM_HTTP_PORT` | port number | `8080` |
| `HELIUM_GRPC_PORT` | port number | `50050` |
| `HELIUM_LOG_LEVEL` | `info`, `debug`, `warn`, `error` | `info` |

### Gold

| Variable | Format | Example |
|----------|--------|---------|
| `GOLD_DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | `postgres://snowflake:password@db.example.com:5432/gold?sslmode=require` |
| `GOLD_RABBITMQ_URL` | `amqp://user:pass@host:5672` | `amqp://snowflake:password@mq.example.com:5672` |
| `GOLD_JWT_SECRET` | any string (min 32 chars recommended) | `your-secret-key-here` |
| `GOLD_JWT_ISSUER` | any string | `snowflake` |
| `GOLD_HTTP_PORT` | port number | `8083` |
| `GOLD_GRPC_PORT` | port number | `50053` |
| `GOLD_LOG_LEVEL` | `info`, `debug`, `warn`, `error` | `info` |

### Carbon (Scorer)

| Variable | Format | Example | Notes |
|----------|--------|---------|-------|
| `CARBON_DATABASE_URL` | `jdbc:postgresql://host:port/db` | `jdbc:postgresql://db.example.com:5432/carbon` | ⚠️ JDBC format (not postgres://) |
| `CARBON_RABBITMQ_URL` | `amqp://user:pass@host:5672` | `amqp://snowflake:password@mq.example.com:5672` | |
| `CARBON_GRPC_HOST` | hostname or IP | `snowflake-helium` (local) or `helium.example.com` (prod) | Helium service address |
| `CARBON_HTTP_PORT` | port number | `8081` | |
| `CARBON_LOG_LEVEL` | `INFO`, `DEBUG`, `WARN`, `ERROR` | `INFO` | Java logging levels |

### Oxygen (Loan)

| Variable | Format | Example | Notes |
|----------|--------|---------|-------|
| `OXYGEN_DATABASE_URL` | `Host=host;Database=db;Username=user;Password=pass` | `Host=db.example.com;Database=oxygen;Username=snowflake;Password=password` | ⚠️ .NET/Npgsql format |
| `OXYGEN_GRPC_URL` | `http://host:port` | `http://snowflake-helium:50050` (local) or `http://helium.example.com:50050` (prod) | Helium gRPC endpoint |
| `OXYGEN_JWT_SECRET` | any string (min 32 chars recommended) | `your-secret-key-here` | |
| `OXYGEN_HTTP_PORT` | port number | `8082` | |
| `OXYGEN_LOG_LEVEL` | `Information`, `Debug`, `Warning`, `Error` | `Information` | .NET logging levels |

---

## Local Development

### Option A: Use YAML/properties files (no env vars)

Run services **without docker-compose** (run locally via IDE or `go run`/`dotnet run`/`mvn`):
- helium, gold read `config/local.yml`
- carbon reads `application-dev.properties`
- oxygen reads `appsettings.json`
- All use reasonable local defaults

Run postgres and rabbitmq via docker-compose if needed:
```bash
docker compose up postgres rabbitmq
```

### Option B: Test full docker-compose locally

Create a `.gitignored` `.env` file at the repo root:
```bash
# .env (gitignored)
HELIUM_DATABASE_URL=postgres://snowflake:snowflake@postgres:5432/helium?sslmode=disable
HELIUM_RABBITMQ_URL=amqp://snowflake:snowflake@rabbitmq:5672
HELIUM_JWT_SECRET=dev-secret-key
# ... repeat for GOLD_, CARBON_, OXYGEN_
```

Then:
```bash
docker compose up
```

Docker Compose substitutes variables from `.env` automatically.

---

## Production Deployment (Coolify)

1. **Commit this file and any code changes** to the repository
2. **In Coolify:**
   - Create a new Docker Compose project pointing to your repo
   - Add all `{SERVICE}_*` environment variables via Coolify's UI (use the tables above)
   - Deploy

Coolify automatically:
- Substitutes the variables into `docker-compose.yml`
- Passes them to each container
- Services read YAML defaults, then apply env var overrides

---

## Format Notes

### Database URLs

- **helium, gold** (Go/GORM): `postgres://user:pass@host:5432/db?sslmode=disable`
- **carbon** (Spring Boot/JDBC): `jdbc:postgresql://host:5432/db` — **different format!**
- **oxygen** (.NET/Npgsql): `Host=host;Database=db;Username=user;Password=pass` — **different format!**

This is intentional — each service uses its native driver's format. When setting `CARBON_DATABASE_URL` in Coolify, use JDBC format. For `OXYGEN_DATABASE_URL`, use Npgsql format.

---

## Future Unification (Optional)

If you want all services to accept `DATABASE_URL` in a single format (e.g., `postgres://...`):
- carbon: add a Spring Boot configuration class to parse postgres:// format and convert to JDBC
- oxygen: configure Npgsql to accept postgres:// URIs natively (minor appsettings change)

This is not required for the MVP — the current format-per-service approach works fine once documented.

---

## Checklist

- [ ] Add env var support to `helium/pkg/config/config.go`
- [ ] Add env var support to `gold/pkg/config/config.go`
- [ ] Add `environment:` blocks to all 4 services in `docker-compose.yml`
- [ ] Test locally (with `.env` or without)
- [ ] Commit all changes
- [ ] Set up Coolify project with all `{SERVICE}_*` variables populated
- [ ] Deploy and verify services start with correct config
