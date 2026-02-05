# --- Checks for required tools ---
check-docker:
	@command -v docker > /dev/null 2>&1 || { echo "Docker is not installed. Please install Docker."; exit 1; }

check-dotnet:
	@command -v dotnet > /dev/null 2>&1 || { echo ".NET is not installed. Please install .NET."; exit 1; }

check-node:
	@command -v node > /dev/null 2>&1 || { echo "Node is not installed. Please install Node."; exit 1; }

# --- Database migration targets ---
migrate-oxygen: check-dotnet
	dotnet ef database update --project oxygen/Loan.API

# --- Docker and service start targets ---
start: check-docker check-dotnet
	docker-compose up -d
	make migrate-oxygen

# --- OpenAPI merge related ---
openapi: check-node
	@mkdir -p docs
	npx openapi-merge-cli --config docs/openapi-merge.json
	@echo "✓ OpenAPI documentation generated at docs/openapi.json"