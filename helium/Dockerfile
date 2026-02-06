# Builder
FROM golang:1.22.4-alpine AS builder

WORKDIR /app

# Install curl
RUN apk --no-cache add curl make

# Set the version of protoc you want to install
ENV PROTOC_VERSION=27.1

# Download and install protoc
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip \
    && unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local \
    && rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Install protoc-gen-go and protoc-gen-go-grpc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1 \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Generate protobuf code
RUN make generate
RUN go build -o main src/cmd/server.go

# Runner
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/swagger.json .
ENV APP_CONF=/app/config/prod.yml

EXPOSE 8080
EXPOSE 50050

CMD ["./main"]
