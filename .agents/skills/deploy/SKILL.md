---
name: deploy
description: Use when deploying services, configuring env vars for production, setting up Coolify, writing CI/CD workflows, troubleshooting service startup, or needing database URL format per service.
---

# Deploy

One config file per service. Env vars override file defaults.

## Services

| Service | Stack | Config File | Env Var Convention |
|---------|-------|-------------|-------------------|
| helium | Go | `config/local.yml` | dots→underscores, `{SERVICE}_*` prefix |
| gold | Go | `config/local.yml` | dots→underscores, `{SERVICE}_*` prefix |
| carbon | Java | `application.properties` | `SPRING_*` prefix |
| oxygen | .NET | `appsettings.json` | CamelCase, double-underscore |

## Env Var Examples

- **helium:** `HELIUM_DATABASE_URL`, `HELIUM_RABBITMQ_URL`, `HELIUM_JWT_SECRET`
- **gold:** `GOLD_DATABASE_URL`, `GOLD_JWT_SECRET`
- **carbon:** `SPRING_DATASOURCE_URL` (needs JDBC format), `LOGGING_LEVEL_ROOT`
- **oxygen:** `ConnectionStrings__DefaultConnection`, `Logging__LogLevel__Default`

Config path override: `APP_CONF=/path/to/config.yml` (Go services only).

## CI/CD Pipeline

Each service independent:
1. Push `main` → `.github/workflows/publish-<service>.yml`
2. Build Docker image → push `ghcr.io/<repo>/<service>:prod`
3. Call Coolify webhook `COOLIFY_<SERVICE>_WEBHOOK` → redeploy
4. Coolify pulls image, restarts container with configured env vars

No shared docker-compose in prod. No nginx in prod.

## Config Priority

1. Config file defaults (lowest)
2. Environment variables
3. Explicit env var bindings (highest)

## Database URL Formats

Per-service native driver format. Use correct format for Coolify env vars:

| Stack | Driver | Format |
|-------|--------|--------|
| Go | GORM | `postgres://user:pass@host:5432/db?sslmode=disable` |
| Java | JDBC | `jdbc:postgresql://host:5432/db` |
| .NET | Npgsql | `Host=host;Database=db;Username=user;Password=pass` |
