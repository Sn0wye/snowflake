# --- Checks for required tools ---
check-docker:
	@command -v docker > /dev/null 2>&1 || { echo "Docker is not installed. Please install Docker."; exit 1; }

check-dotnet:
	@command -v dotnet > /dev/null 2>&1 || { echo ".NET is not installed. Please install .NET."; exit 1; }

check-node:
	@command -v node > /dev/null 2>&1 || { echo "Node is not installed. Please install Node."; exit 1; }

check-go:
	@command -v go > /dev/null 2>&1 || { echo "Go is not installed. Please install Go."; exit 1; }

check-swag:
	@command -v swag > /dev/null 2>&1 || { echo "swag is not installed. Please install swag (go install github.com/swaggo/swag/cmd/swag@latest)."; exit 1; }

# --- Database migration targets ---
migrate-oxygen: check-dotnet
	dotnet ef database update --project oxygen/Loan.API

# --- Docker and service start targets ---
start: check-docker check-dotnet
	docker-compose up -d
	make migrate-oxygen

# --- OpenAPI generation targets (individual services) ---
docs-helium: check-swag
	$(MAKE) -C helium docs

docs-gold: check-swag
	$(MAKE) -C gold docs

# --- OpenAPI generation (all services) ---
docs-generate: docs-helium docs-gold
	@echo "✓ All service OpenAPI specs generated"

# --- OpenAPI merge related ---
openapi: docs-generate check-node
	npx openapi-merge-cli --config docs/openapi-merge.json
	@echo "✓ OpenAPI documentation merged at docs/openapi.json"