# Deployment Configuration Plan

## Overview

All microservices use **committed dev/local config files as defaults** and support **environment variable overrides**. Config files are source of truth; env vars provide runtime overrides. No `prod.yml` or `application-prod.properties` files needed — single config file per service applies everywhere, with env vars enabling environment-specific customization.

---

## Architecture

**All environments (local dev, docker-compose, production):**

- helium: reads `config/local.yml` (default), overridden by env vars
- gold: reads `config/local.yml` (default), overridden by env vars
- carbon: reads `application.properties`, overridden by Spring Boot env var conventions
- oxygen: reads `appsettings.json` (or `appsettings.Development.json` in dev), overridden by .NET configuration conventions

---

## Code Changes Required

### helium & gold (`pkg/config/config.go`)

**Current implementation:** ✅ Already configured correctly.

Services load config via:

1. `APP_CONF` env var (if set) — allows runtime config file override
2. Default: `config/local.yml` relative to working directory

Env var override support already enabled:

```go
// Automatic env var binding with underscore replacement (dots → underscores)
conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
conf.AutomaticEnv()

// Explicit conventional bindings — only applied if env var is non-empty
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

- `DATABASE_URL`, `RABBITMQ_URL`, `JWT_SECRET` override YAML defaults
- All other YAML keys overridable via uppercased env vars with dots replaced by underscores
  - Example: `HTTP_PORT` → `http.port`, `LOG_LOG_LEVEL` → `log.log_level`

### carbon (Spring Boot)

**Current implementation:** ✅ Natively supports env var overrides.

Spring Boot automatically binds env vars to properties using convention:

- `SPRING_*` prefixed vars map to `spring.*` properties
- Example: `SPRING_DATASOURCE_URL` → `spring.datasource.url`

No code changes needed.

### oxygen (.NET)

**Current implementation:** ✅ Natively supports env var overrides.

.NET Configuration uses double-underscore notation:

- `ConnectionStrings__DefaultConnection` → `ConnectionStrings:DefaultConnection`
- `Logging__LogLevel__Default` → `Logging:LogLevel:Default`

No code changes needed.

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

| Variable              | Format                                              | Example                                                                    | Default                                                                |
| --------------------- | --------------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `HELIUM_DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | `postgres://snowflake:password@db.example.com:5432/helium?sslmode=require` | `postgres://snowflake:snowflake@127.0.0.1:5432/helium?sslmode=disable` |
| `HELIUM_RABBITMQ_URL` | `amqp://user:pass@host:5672`                        | `amqp://snowflake:password@mq.example.com:5672`                            | `amqp://snowflake:snowflake@localhost:5672`                            |
| `HELIUM_JWT_SECRET`   | any string (min 32 chars recommended)               | `your-secret-key-here`                                                     | `super-secret-key`                                                     |
| `HELIUM_JWT_ISSUER`   | any string                                          | `snowflake`                                                                | `snowflake`                                                            |
| `HELIUM_HTTP_PORT`    | port number                                         | `8080`                                                                     | `8080`                                                                 |
| `HELIUM_GRPC_PORT`    | port number                                         | `50050`                                                                    | `50050`                                                                |
| `HELIUM_LOG_LEVEL`    | `info`, `debug`, `warn`, `error`                    | `info`                                                                     | `info`                                                                 |

### Gold

| Variable            | Format                                              | Example                                                                  | Default                                                              |
| ------------------- | --------------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------- |
| `GOLD_DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | `postgres://snowflake:password@db.example.com:5432/gold?sslmode=require` | `postgres://snowflake:snowflake@127.0.0.1:5432/gold?sslmode=disable` |
| `GOLD_RABBITMQ_URL` | `amqp://user:pass@host:5672`                        | `amqp://snowflake:password@mq.example.com:5672`                          | `amqp://snowflake:snowflake@localhost:5672`                          |
| `GOLD_JWT_SECRET`   | any string (min 32 chars recommended)               | `your-secret-key-here`                                                   | `super-secret-key`                                                   |
| `GOLD_JWT_ISSUER`   | any string                                          | `snowflake`                                                              | `snowflake`                                                          |
| `GOLD_HTTP_PORT`    | port number                                         | `8083`                                                                   | `8083`                                                               |
| `GOLD_GRPC_PORT`    | port number                                         | `50053`                                                                  | `50053`                                                              |
| `GOLD_LOG_LEVEL`    | `info`, `debug`, `warn`, `error`                    | `info`                                                                   | `info`                                                               |

### Carbon (Scorer)

| Variable                    | Format                           | Example                                        | Default                                   | Notes                         |
| --------------------------- | -------------------------------- | ---------------------------------------------- | ----------------------------------------- | ----------------------------- |
| `SPRING_DATASOURCE_URL`     | `jdbc:postgresql://host:port/db` | `jdbc:postgresql://db.example.com:5432/carbon` | `jdbc:postgresql://localhost:5432/carbon` | JDBC format (not postgres://) |
| `SPRING_RABBITMQ_ADDRESSES` | `host:port` or `amqp://...`      | `mq.example.com:5672`                          | `localhost:5672`                          |                               |
| `SPRING_GRPC_HOST`          | hostname or IP                   | `helium.example.com`                           | `127.0.0.1`                               | Helium service address        |
| `SERVER_PORT`               | port number                      | `8081`                                         | `8081`                                    |                               |
| `LOGGING_LEVEL_ROOT`        | `INFO`, `DEBUG`, `WARN`, `ERROR` | `INFO`                                         | `DEBUG`                                   | Java logging levels           |

### Oxygen (Loan)

| Variable                               | Format                                              | Example                                                                    | Default                                                                | Notes                |
| -------------------------------------- | --------------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------- | -------------------- |
| `ConnectionStrings__DefaultConnection` | `Host=host;Database=db;Username=user;Password=pass` | `Host=db.example.com;Database=oxygen;Username=snowflake;Password=password` | `Host=localhost;Database=oxygen;Username=snowflake;Password=snowflake` | .NET/Npgsql format   |
| `ConnectionStrings__UsersGrpc`         | `http://host:port`                                  | `http://helium.example.com:50050`                                          | `http://127.0.0.1:50050`                                               | Helium gRPC endpoint |
| `Security__Jwt__SecretKey`             | any string (min 32 chars recommended)               | `your-secret-key-here`                                                     | `super-secret-key`                                                     |                      |
| `ASPNETCORE_HTTP_PORTS`                | port number                                         | `8082`                                                                     | `8082`                                                                 |                      |
| `Logging__LogLevel__Default`           | `Information`, `Debug`, `Warning`, `Error`          | `Information`                                                              | `Information`                                                          | .NET logging levels  |

---

## Local Development

### Option A: Use config files (no env vars)

Run services **without docker-compose** (run locally via IDE or `go run`/`dotnet run`/`mvn`):

- helium, gold read `config/local.yml` (defaults)
- carbon reads `application.properties` (defaults)
- oxygen reads `appsettings.Development.json` (defaults)

Run postgres and rabbitmq via docker-compose if needed:

```bash
docker compose up postgres rabbitmq
```

### Option B: Test full docker-compose locally

Create a `.gitignored` `.env` file at repo root with service-prefixed vars:

```bash
# .env (gitignored)
HELIUM_DATABASE_URL=postgres://snowflake:snowflake@postgres:5432/helium?sslmode=disable
HELIUM_RABBITMQ_URL=amqp://snowflake:snowflake@rabbitmq:5672
HELIUM_JWT_SECRET=dev-secret-key
# ... repeat for GOLD_, SPRING_, ConnectionStrings__, etc.
```

Then:

```bash
docker compose up
```

Docker Compose substitutes variables from `.env` automatically.

**Note:** `docker-compose.yml` already has sensible defaults (e.g., `${HELIUM_DATABASE_URL:-}`) so unset vars fall back to the service config file defaults.

---

## Production Deployment (Coolify)

1. **Ensure all code is committed** to the repository
2. **In Coolify:**
   - Create a new Docker Compose project pointing to your repo
   - Add all `{SERVICE}_*` and Spring/Logging env vars via Coolify's UI (use the tables above)
   - Deploy

Coolify automatically:

- Substitutes the variables into `docker-compose.yml`
- Passes them to each container
- Services load config file defaults, then apply env var overrides

---

## Configuration Priority (lowest to highest)

1. **Config file defaults** (`local.yml`, `application.properties`, `appsettings.json`)
2. **Environment variables** (override file defaults)
3. **Explicit env var bindings** (e.g., `DATABASE_URL` → `db.connectionString`)

---

## Format Notes

### Database URLs

- **helium, gold** (Go/GORM): `postgres://user:pass@host:5432/db?sslmode=disable`
- **carbon** (Spring Boot/JDBC): `jdbc:postgresql://host:5432/db` — **different format!**
- **oxygen** (.NET/Npgsql): `Host=host;Database=db;Username=user;Password=pass` — **different format!**

This is intentional — each service uses its native driver's format. When setting `SPRING_DATASOURCE_URL` in Coolify, use JDBC format. For `ConnectionStrings__DefaultConnection`, use Npgsql format.

---
