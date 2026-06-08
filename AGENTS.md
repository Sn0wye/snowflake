- All interactions, commit messages: concise. Sacrifice grammar for brevity.

## Config & Env Vars

One config file per service. Env vars override file defaults.

- helium/gold: `config/local.yml`
- carbon: Reads `application.properties`
- oxygen: Reads `appsettings.json` (or `.Development.json` in dev)

For deploy, env vars, CI/CD, Coolify config: use `deploy` skill.

## Docker Compose (Local Dev Only)

- Env var substitution: `${SERVICE_VAR:-default}` pattern throughout
- `.env` file (gitignored) for local dev overrides
- Services need cwd = service root, not repo root

## Testing & Verification

- Go: `go test ./...` (from service root)
- Java: `mvn test` (from `./carbon`)
- C#: `dotnet test` (from `./oxygen`)
