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
	swag fmt -g src/cmd/server.go -d ./src

docs: docs-fmt
	swag init -g src/cmd/server.go --ot json -o ./
	make openapi-gen
