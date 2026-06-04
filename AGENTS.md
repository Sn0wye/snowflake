# AGENTS.md — Snowflake Microservices

Microservices monorepo: **4 services** (Go, Java, C#) + **Docker Compose** + **PostgreSQL/RabbitMQ**.

## Services & Ownership

| Service | Path | Lang | Framework | Ports | Key Files |
|---------|------|------|-----------|-------|-----------|
| **helium** | `./helium` | Go | Fiber | 8080 HTTP, 50050 gRPC | `pkg/config/config.go`, `src/cmd/server.go` |
| **gold** | `./gold` | Go | Fiber | 8083 HTTP | `pkg/config/config.go`, `src/cmd/server.go` |
| **carbon** | `./carbon` | Java | Spring Boot | 8081 HTTP | `src/main/resources/application.properties` |
| **oxygen** | `./oxygen` | C# | .NET 8 | 8082 HTTP | `Oxygen.API/appsettings.json` |

## Quick Commands

```bash
# Start everything
make start

# Start only postgres + rabbitmq
docker compose up postgres rabbitmq

# Regenerate all OpenAPI docs (merges all service specs)
make openapi

# Generate Go proto code (helium/gold)
cd helium && make generate  # or `cd gold && make generate`

# Migrate .NET database
make migrate-oxygen
```

## Configuration & Env Vars

**One config file per service** (no prod configs). Env vars override file defaults:

- **helium/gold**: Read `config/local.yml` by default. `APP_CONF=/path/to/config.yml` overrides.
  - Env var binding: dots → underscores (`HTTP_PORT` → `http.port`)
  - Explicit bindings: `DATABASE_URL`, `RABBITMQ_URL`, `JWT_SECRET`
- **carbon**: Reads `application.properties`. Spring Boot auto-binds `SPRING_*` env vars.
  - Exception: `SPRING_DATASOURCE_URL` expects JDBC format (`jdbc:postgresql://...`), not `postgres://...`
- **oxygen**: Reads `appsettings.json` (or `.Development.json` in dev). .NET uses double-underscores (`ConnectionStrings__DefaultConnection`).

**See DEPLOYMENT.md for full env var reference and local dev setup.**

## Docker Compose (Local Dev Only)

- `docker-compose.yml` is for local development only — not used in production.
- Includes nginx reverse proxy on port 80 (also local dev only).
- Env var substitution: `${SERVICE_VAR:-default}` pattern throughout.
- `.env` file (gitignored) provides local overrides for development.
- **Fixed June 2026**: YAML `depends_on` indentation was broken; now correct.

## Production Deployment

Each service deploys independently via GitHub Actions → ghcr.io → Coolify:

- Push to `main` triggers `.github/workflows/publish-<service>.yml`
- Image pushed to `ghcr.io/<repo>/<service>:prod`
- Coolify webhook (`COOLIFY_<SERVICE>_WEBHOOK`) redeploys the service
- Each service is configured individually in Coolify with its own env vars
- No shared compose file or nginx in production

## Development Gotchas

### Go Services (helium/gold)

- **Config loading**: Always resolves relative to working directory. When running via IDE, set working directory to service root (`./helium` or `./gold`).
- **Hot reload**: Uses Air (`.air.toml`). Binary output: `./tmp/main`.
- **Proto codegen**: Manual step via `make generate` (not automatic on save).
- **Message format**: Both services produce OpenAPI docs + gRPC. Keep `swagger.json` in sync for OpenAPI merging.

### Java Service (carbon)

- **Properties file precedence**: `application.properties` + Spring env var conventions (`SPRING_*`).
- **Database URL format**: JDBC format only (`jdbc:postgresql://...`). Postgres driver won't accept standard URLs.
- **Logging level**: Java convention is uppercase (`DEBUG`, `INFO`, `ERROR`) not lowercase.
- **Queue binding**: RabbitMQ address parsing — use `host:port` or full amqp URI depending on config key.

### C# Service (oxygen)

- **Configuration**: Multiple appsettings files layered (appsettings.json + appsettings.Development.json in dev).
- **EF Migrations**: Live in `Oxygen.Infrastructure`. Always run `make migrate-oxygen` after schema changes.
- **Double-underscore notation**: `Logging__LogLevel__Default` (not `Logging:LogLevel:Default` in env vars).
- **Project structure**: 6 projects (API, DTO, Domain, Infrastructure, Repository, Service, Proto) — changes often span multiple.

## Testing & Verification

**No explicit test commands documented.** Each service likely has tests in `*_test.go`, `**/*Test.java`, `**/*Tests.cs` but no Makefile/script targets yet. Assume:
- Go: `go test ./...` (from service root)
- Java: `mvn test` (from `./carbon`)
- C#: `dotnet test` (from `./oxygen`)

## Messaging & gRPC

- **RabbitMQ**: Shared broker for async events. Gold publishes transaction events; carbon consumes scoring events.
- **gRPC**: helium exposes gRPC on port 50050 (auth service). Carbon & oxygen call it.
- **Proto files**: `helium/proto/` and `gold/proto/`. Regenerate with `make generate` after changes.

## Common Mistakes to Avoid

1. **Config file path**: Don't assume `config/prod.yml` exists — it doesn't. All environments use `local.yml`.
2. **Database URL formats**: Each service uses a different format (postgres://, jdbc:postgresql://, Host=...). Copy from `.env.example`.
3. **gRPC ports vs HTTP ports**: helium gRPC (50050) ≠ carbon HTTP (8081). Check docker-compose for mappings.
4. **Env var naming**: helium/gold use `UPPERCASE_SNAKE`, carbon uses `SPRING_*`, oxygen uses `CamelCase__With__Underscores`.
5. **Working directory in dev**: Go services need working directory = service root, not repo root.
6. **Dockerfile rebuilds**: After code changes, `docker-compose up` won't rebuild. Use `docker-compose up --build` or `docker-compose build`.

## Repository Notes

- Learning project — intentional overengineering across 4 languages + async messaging + gRPC.
- No pre-commit hooks or automatic linting configured.
- OpenAPI specs merged manually via `make openapi` (uses openapi-merge-cli from Node).
- `.opencode/agents/general.md` defines OpenCode subagent behavior for this repo.
