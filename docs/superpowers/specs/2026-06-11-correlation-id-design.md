# Correlation ID Propagation & Structured Logging — Design Spec

**Date:** 2026-06-11  
**Issue:** SNF-31  
**Approach:** Lightweight `X-Correlation-ID` propagation with automatic log injection (Approach B)

---

## Problem

No request correlation or distributed tracing exists across services. It is impossible to trace a single user request through the call chain: `client → helium → gold → rabbitmq → carbon` or `oxygen → carbon → helium (gRPC)`. Logs are unstructured plain text in Carbon and Oxygen, making machine querying impossible.

---

## Goals

1. Propagate a `X-Correlation-ID` (UUID v4) through all service boundaries: HTTP headers, gRPC metadata, and RabbitMQ AMQP headers.
2. Automatically inject `correlation_id` into every log line within a request scope — no per-call developer work required.
3. Emit JSON-structured logs to stdout in all services (prep for Loki/Tempo LGTM stack).

## Non-Goals

- OpenTelemetry trace/span instrumentation (future task).
- nginx/Traefik gateway changes (proxy-agnostic by design).
- Log aggregation infrastructure (separate task).

---

## Shared Constants

| Concern | Value |
|---------|-------|
| HTTP header | `X-Correlation-ID` |
| gRPC metadata key | `x-correlation-id` (lowercase, gRPC convention) |
| AMQP header key | `x-correlation-id` |
| Log field key | `correlation_id` |
| ID format | UUID v4 |

**Generation rule:** If the incoming header/metadata/AMQP header contains a non-empty `X-Correlation-ID`, use it. Otherwise generate a new UUID v4. Never reject a request due to a missing ID. Always echo the used ID back in the response header.

---

## Architecture

```
External client
  → HTTP request (X-Correlation-ID optional)
    → Service middleware: read or generate UUID v4
      → enrich logger with correlation_id field
      → all log calls within request include correlation_id automatically
      → outbound HTTP:   forward X-Correlation-ID header
      → outbound gRPC:   forward as metadata key "x-correlation-id"
      → outbound AMQP:   stamp amqp.Table{"x-correlation-id": id}
    → Consuming service: read from header/metadata/AMQP header → same enrichment
```

---

## Service-by-Service Design

### Helium (Go / Fiber v2)

**Logging:** Create `config/production.yml` (does not exist yet) with `log.encoding: json` and no lumberjack file sink — stdout only (12-factor). The dev `config/local.yml` keeps `encoding: console`.

**New file: `helium/src/middleware/correlation.go`**
- Fiber middleware: reads `X-Correlation-ID` request header, generates UUID v4 if absent.
- Creates a `*zap.Logger` with `zap.String("correlation_id", id)` pre-attached.
- Stores logger in `c.Locals("logger")`.
- Sets `X-Correlation-ID` response header.
- Registered globally before all route groups in `routes.go`.

**Handler convention:** Handlers access the enriched logger via `c.Locals("logger").(*zap.Logger)` instead of the global logger.

**gRPC server (`helium/src/grpc/interceptors.go`):**
- New `CorrelationInterceptor` added before `LoggingInterceptor` in the chain.
- Reads `x-correlation-id` from incoming gRPC metadata.
- Stores in context via a typed context key (`type correlationKey struct{}`).
- `LoggingInterceptor` reads from context to include in gRPC access logs.

**RabbitMQ publish (`helium/pkg/messaging/rabbitmq.go`):**
- Publisher method accepts correlation ID parameter.
- Stamps `amqp091.Table{"x-correlation-id": id}` on `amqp091.Publishing.Headers`.
- Call site extracts ID from Fiber context before publishing.

---

### Gold (Go / Fiber v2)

Symmetric to Helium for HTTP middleware, logger config, and handler convention.

**RabbitMQ consume (`gold/src/cmd/server/main.go` consumer goroutine):**
- Extract `x-correlation-id` from `amqp.Delivery.Headers`.
- Create enriched `*zap.Logger` for that message's processing scope.

**Outbox worker (`gold/src/outbox/worker.go`):**
- Outbox entries must carry `correlation_id` column (or store in event payload).
- Worker reads it and forwards via AMQP headers when re-publishing `transaction.*` events.

**New file: `gold/pkg/middleware/correlation.go`** — identical pattern to Helium.

---

### Carbon (Java / Spring Boot 3.2.5)

**Logging — full implementation:**
- Add `net.logstash.logback:logstash-logback-encoder` to `pom.xml`.
- Add `logback-spring.xml` with `ConsoleAppender` using `LogstashEncoder` — JSON to stdout.
- Replace all `System.out.println` calls with `@Slf4j` logger (`log.info`, `log.error`).
- Use SLF4J structured args (`log.info("msg: {}", value)`) not string concatenation.

**Correlation ID — via SLF4J MDC:**

**New class: `CorrelationIdInterceptor`** (separate from `BearerTokenInterceptor`, single responsibility):
- Reads `X-Correlation-ID` from `HttpServletRequest`, generates UUID if absent.
- Calls `MDC.put("correlation_id", id)`.
- Sets `X-Correlation-ID` on `HttpServletResponse`.
- Clears MDC key in `afterCompletion`.
- Registered before `BearerTokenInterceptor` in `WebMvcConfig`.
- `LogstashEncoder` automatically includes all MDC entries in every JSON log line.

**RabbitMQ consume (`ScoreConsumer.java`):**
- Extract `x-correlation-id` from `message.getMessageProperties().getHeaders()`.
- `MDC.put("correlation_id", id)` before processing, clear in finally block.
- Replace `System.out.println` with `log.info` / `log.error`.

---

### Oxygen (C# / ASP.NET Core 8)

**Logging — add Serilog:**
- Add packages: `Serilog.AspNetCore`, `Serilog.Sinks.Console`, `Serilog.Formatting.Compact`.
- Replace `builder.Logging` in `Program.cs` with `UseSerilog(...)` configured with `new CompactJsonFormatter()`.
- Replace `Console.WriteLine` in `Program.cs` and `UsersGRPCAdapter.cs` with injected `ILogger<T>`.
- All existing `ILogger<T>` call sites remain unchanged — Serilog wires under MEL.

**Correlation ID — via Serilog LogContext:**

`ExceptionHandlingMiddleware` (already first in pipeline) extended to:
- Read `X-Correlation-ID` from `HttpContext.Request.Headers`, generate UUID if absent.
- Set `HttpContext.Response.Headers["X-Correlation-ID"]`.
- Store in `HttpContext.Items["correlation_id"]` for use by outbound adapters.
- Open `LogContext.PushProperty("correlation_id", id)` — `using` scope enriches all Serilog calls for request lifetime.

**Outbound HTTP (`CreditScoreAdapter.cs` → Carbon):**
- Read correlation ID from `IHttpContextAccessor.HttpContext.Items["correlation_id"]`.
- Forward as `X-Correlation-ID` header on `HttpClient` request.
- Register `IHttpContextAccessor` in DI (`Program.cs`).

**Outbound gRPC (`UsersGRPCAdapter.cs` → Helium `UserService`):**
- Read correlation ID from `IHttpContextAccessor`.
- Stamp into `CallOptions` metadata: `Metadata { { "x-correlation-id", id } }`.
- Also forward `Authorization` token as gRPC metadata (fixes latent auth bug — Helium's `AuthInterceptor` requires it for all `UserService` methods; currently Oxygen sends nothing).

---

## Testing Strategy

- **Unit:** Middleware/interceptor logic — verify ID generation when header absent, ID passthrough when present, response header set correctly.
- **Integration:** Per-service — make an HTTP request with a known `X-Correlation-ID`, assert the value appears in log output JSON.
- **E2E (manual):** `docker-compose up`, send a request with a fixed correlation ID, grep logs across services for that ID to confirm propagation through the full chain.

---

## Files Modified / Created

| Service | File | Change |
|---------|------|--------|
| helium | `src/middleware/correlation.go` | new |
| helium | `src/routes/routes.go` | register middleware |
| helium | `src/grpc/interceptors.go` | add CorrelationInterceptor |
| helium | `pkg/messaging/rabbitmq.go` | stamp AMQP header |
| helium | `config/production.yml` | `encoding: json`, remove lumberjack |
| gold | `pkg/middleware/correlation.go` | new |
| gold | `src/cmd/server/main.go` | register middleware, AMQP consume extract |
| gold | `src/outbox/worker.go` | forward correlation_id on re-publish |
| gold | `pkg/messaging/rabbitmq.go` | stamp AMQP header |
| gold | `config/production.yml` | `encoding: json`, remove lumberjack |
| carbon | `pom.xml` | add logstash-logback-encoder |
| carbon | `src/main/resources/logback-spring.xml` | new — JSON console appender |
| carbon | `interceptors/CorrelationIdInterceptor.java` | new |
| carbon | `config/WebMvcConfig.java` | register interceptor |
| carbon | `consumers/ScoreConsumer.java` | MDC from AMQP header, replace println |
| oxygen | `Oxygen.API.csproj` | add Serilog packages |
| oxygen | `Oxygen.API/Program.cs` | UseSerilog, replace Console.WriteLine |
| oxygen | `Oxygen.API/Middleware/ExceptionHandlingMiddleware.cs` | LogContext + correlation ID |
| oxygen | `Oxygen.Infrastructure/Adapters/CreditScoreAdapter.cs` | forward header |
| oxygen | `Oxygen.Infrastructure/Adapters/UsersGRPCAdapter.cs` | stamp gRPC metadata, fix auth bug |
