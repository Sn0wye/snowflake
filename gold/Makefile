# -- OpenAPI Gen --
INPUT_FILE = ./swagger.json
OUTPUT_FILE = ./openapi.json
URL = https://converter.swagger.io/api/convert
openapi-gen:
	@echo "Sending request to $(URL) with $(INPUT_FILE)..."
	@curl -s -X POST $(URL) \
		-H "Content-Type: application/json" \
		-d @$(INPUT_FILE) \
		-o $(OUTPUT_FILE)
	@echo "Response has been written to $(OUTPUT_FILE)"

docs-fmt:
	swag fmt -g src/cmd/server/main.go -d ./src

docs: docs-fmt
	swag init -g src/cmd/server/main.go --ot json -o ./
	make openapi-gen

# -- Migrations --
MIGRATIONS_DIR := src/migration/migrations
DATABASE_URL ?= postgres://snowflake:snowflake@127.0.0.1:5432/gold?sslmode=disable

migrate-up:
	@DATABASE_URL=$(DATABASE_URL) go run src/cmd/migrate/main.go up

migrate-down:
	@DATABASE_URL=$(DATABASE_URL) go run src/cmd/migrate/main.go down

migrate-create:
	@[ -n "$(NAME)" ] || (echo "Usage: make migrate-create NAME=add_foo_column" && exit 1)
	@go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
		create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
