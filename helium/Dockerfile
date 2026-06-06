# Builder
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o main src/cmd/server/main.go
RUN go build -ldflags="-s -w" -o migrate src/cmd/migrate/main.go

# Runner
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/config ./config

EXPOSE 8080
EXPOSE 50050

CMD ["./main"]
