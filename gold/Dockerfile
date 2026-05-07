# Builder
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

RUN apk --no-cache add make

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main src/cmd/server.go

# Runner
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/swagger.json .

EXPOSE 8083

CMD ["./main"]
