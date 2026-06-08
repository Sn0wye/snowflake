- In all interactions and commit messages, be extremely concise and sacrifice grammar for the sake of concision. I

## Configuration & Env Vars

One config file per service. Env vars override file defaults.

- helium/gold: `config/local.yml`
- carbon: Reads `application.properties`.
- oxygen: Reads `appsettings.json` (or `.Development.json` in dev).

**See DEPLOYMENT.md for full env var reference and local dev setup.**

## Docker Compose (Local Dev Only)

- Env var substitution: `${SERVICE_VAR:-default}` pattern throughout.
- `.env` file (gitignored) provides local overrides for development.
- Working directory in dev: Services need cwd = service root, not repo root.

## Testing & Verification

- Go: `go test ./...` (from service root)
- Java: `mvn test` (from `./carbon`)
- C#: `dotnet test` (from `./oxygen`)
