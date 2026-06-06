# AGENTS.md — Carbon (Credit Score Microservice)

Spring Boot 3.2.5 / Java 21 / PostgreSQL / RabbitMQ / gRPC

## Build & Run

```bash
# Build (compile + proto)
./mvnw compile

# Run (needs PostgreSQL + RabbitMQ running)
./mvnw spring-boot:run

# Package as JAR
./mvnw package -DskipTests
```

## Testing

```bash
# Run all tests (requires no external dependencies)
./mvnw test

# Run a single test class
./mvnw test -Dtest=CreditScoreServiceTest

# Run with mvnw (uses Maven wrapper — no local Maven install needed)
```

Tests use:
- **H2 in-memory database** (replaces PostgreSQL, auto-configured via `src/test/resources/application.properties`)
- **RabbitMQ listeners disabled** (`spring.rabbitmq.listener.simple.auto-startup=false`)
- **OpenAPI export disabled** (`openapi.export.enabled=false`)
- No gRPC server needed (plain unit tests)

31 tests: 1 `@SpringBootTest` context load + 30 credit score calculation tests.

## Common Issues

| Problem | Fix |
|---------|-----|
| `BUILD FAILURE` with connection refused | Run `./mvnw clean test` (stale `target/` may have old class files) |
| PostgreSQL connection error in tests | Ensure `src/test/resources/application.properties` is present (overrides main DB config) |
| `Could not resolve placeholder 'spring.rabbitmq.host'` | Test props must include dummy rabbitmq values (RabbitConfig reads them via `@Value`) |
| RabbitMQ producer test fails with `Connection refused` | `ConnectionFactory` still points at localhost:5672. Listeners won't connect (auto-startup=false), but any test that triggers a publish (e.g. testing a producer bean directly) will attempt a real connection. Mock `RabbitTemplate` or exclude `RabbitConfig` in such tests. |
| gRPC channel fails in tests that load `GrpcAuthService` | `spring.grpc.host`/`spring.grpc.port` are set but no gRPC server runs in tests. `GrpcAuthService` eagerly creates a `ManagedChannel` in its constructor — any test that autowires it will get a bean with a dead channel. Current tests don't touch it, but if you write a test that depends on auth beans, mock or exclude `GrpcAuthService`. |

## Proto Codegen

Proto files compile automatically during `mvn compile` via protobuf-maven-plugin.

```bash
# Manual: generate from src/main/proto/*.proto
./mvnw protobuf:compile protobuf:compile-custom
```

Generated sources land in `target/generated-sources/protobuf/`.

## Configuration

- **Main**: `src/main/resources/application.properties` (PostgreSQL, RabbitMQ, gRPC)
- **Test**: `src/test/resources/application.properties` (H2, mocked RabbitMQ)
- Env var overrides: `SPRING_DATASOURCE_URL=jdbc:postgresql://...` etc.
