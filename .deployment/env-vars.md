# Environment Variables Reference

Full environment variable specification for all services. Use `.env.example` as a template for local development.

## Helium

| Variable | Format | Example | Default |
|----------|--------|---------|---------|
| `HELIUM_DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | `postgres://snowflake:password@db.example.com:5432/helium?sslmode=require` | `postgres://snowflake:snowflake@127.0.0.1:5432/helium?sslmode=disable` |
| `HELIUM_RABBITMQ_URL` | `amqp://user:pass@host:5672` | `amqp://snowflake:password@mq.example.com:5672` | `amqp://snowflake:snowflake@localhost:5672` |
| `HELIUM_JWT_SECRET` | any string (min 32 chars recommended) | `your-secret-key-here` | `super-secret-key` |
| `HELIUM_JWT_ISSUER` | any string | `snowflake` | `snowflake` |
| `HELIUM_HTTP_PORT` | port number | `8080` | `8080` |
| `HELIUM_GRPC_PORT` | port number | `50050` | `50050` |
| `HELIUM_LOG_LEVEL` | `info`, `debug`, `warn`, `error` | `info` | `info` |

## Gold

| Variable | Format | Example | Default |
|----------|--------|---------|---------|
| `GOLD_DATABASE_URL` | `postgres://user:pass@host:5432/db?sslmode=disable` | `postgres://snowflake:password@db.example.com:5432/gold?sslmode=require` | `postgres://snowflake:snowflake@127.0.0.1:5432/gold?sslmode=disable` |
| `GOLD_RABBITMQ_URL` | `amqp://user:pass@host:5672` | `amqp://snowflake:password@mq.example.com:5672` | `amqp://snowflake:snowflake@localhost:5672` |
| `GOLD_JWT_SECRET` | any string (min 32 chars recommended) | `your-secret-key-here` | `super-secret-key` |
| `GOLD_JWT_ISSUER` | any string | `snowflake` | `snowflake` |
| `GOLD_HTTP_PORT` | port number | `8083` | `8083` |
| `GOLD_GRPC_PORT` | port number | `50053` | `50053` |
| `GOLD_LOG_LEVEL` | `info`, `debug`, `warn`, `error` | `info` | `info` |

## Carbon (Scorer)

| Variable | Format | Example | Default | Notes |
|----------|--------|---------|---------|-------|
| `SPRING_DATASOURCE_URL` | `jdbc:postgresql://host:port/db` | `jdbc:postgresql://db.example.com:5432/carbon` | `jdbc:postgresql://localhost:5432/carbon` | JDBC format (not postgres://) |
| `SPRING_RABBITMQ_ADDRESSES` | `host:port` or `amqp://...` | `mq.example.com:5672` | `localhost:5672` | |
| `SPRING_GRPC_HOST` | hostname or IP | `helium.example.com` | `127.0.0.1` | Helium service address |
| `SERVER_PORT` | port number | `8081` | `8081` | |
| `LOGGING_LEVEL_ROOT` | `INFO`, `DEBUG`, `WARN`, `ERROR` | `INFO` | `DEBUG` | Java logging levels |

## Oxygen (Loan)

| Variable | Format | Example | Default | Notes |
|----------|--------|---------|---------|-------|
| `ConnectionStrings__DefaultConnection` | `Host=host;Database=db;Username=user;Password=pass` | `Host=db.example.com;Database=oxygen;Username=snowflake;Password=password` | `Host=localhost;Database=oxygen;Username=snowflake;Password=snowflake` | .NET/Npgsql format |
| `ConnectionStrings__UsersGrpc` | `http://host:port` | `http://helium.example.com:50050` | `http://127.0.0.1:50050` | Helium gRPC endpoint |
| `Security__Jwt__SecretKey` | any string (min 32 chars recommended) | `your-secret-key-here` | `super-secret-key` | |
| `ASPNETCORE_HTTP_PORTS` | port number | `8082` | `8082` | |
| `Logging__LogLevel__Default` | `Information`, `Debug`, `Warning`, `Error` | `Information` | `Information` | .NET logging levels |
