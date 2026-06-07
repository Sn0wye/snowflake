# Snowflake

This repository serves as a learning project to create and deploy multiple microservices using different technologies and languages, with a focus on enabling communication between them.

TLDR; There will be A LOT of overengineering, but that's the point. The goal is to learn and experiment with different technologies and patterns, not to create a production-ready application.

## Services

| **Service Name** | **Responsibility**             | **Description**                                                                                                                                            | **Language** | **Framework** | **Ports**                 |
| ---------------- | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------------- | ------------------------- |
| Helium           | User authentication management | Helium is lightweight and foundational, just like authentication, which provides the basis for all interactions.                                           | Go           | Fiber         | 8080 (HTTP), 50050 (gRPC) |
| Carbon           | User scoring and evaluation    | Carbon is a fundamental building block, symbolizing the core calculations and essential scoring functions.                                                 | Java         | Spring Boot   | 8081 (HTTP)               |
| Oxygen           | Loan operations and servicing  | Oxygen is life-sustaining, representing the essential and enabling nature of loans in fostering growth and support.                                        | C#           | .NET          | 8082 (HTTP)               |
| Gold             | Payment and balance management | Gold is the universal store of value, representing the financial core that powers instant FLAKE transfers and ledger-based balance tracking between users. | Go           | Fiber         | 8083 (HTTP)               |

## Infrastructure

- **Docker**: Containerization platform
- **Docker Compose**: Orchestration tool for managing multi-container Docker applications
- **PostgreSQL**: Database for storing application data
- **RabbitMQ**: Message broker for asynchronous communication between microservices

## Behind the Decisions

This project is a learning experience where I explore different technologies and patterns. Here's a breakdown of the key decisions:

- **Go & Fiber**: Go is a fast, lightweight language I've been learning and experimenting with. It's ideal for microservices, and its rich ecosystem simplifies development. Fiber, a minimalistic, high-performance web framework, powers both Helium (auth) and Gold (payments), ensuring fast and efficient HTTP servers for both services.

- **Java & Spring Boot**: Leveraging my experience with Java, I chose it for the Scorer service, particularly for interacting with RabbitMQ, as its official client is written in Java. Spring Boot was a natural fit for its extensive ecosystem, making API development and queue consumption seamless.

- **C# & .NET**: As one of my favorite languages, C# was my choice for the Loan service. .NET complements it perfectly for building robust APIs.

- **PostgreSQL**: I selected PostgreSQL for its popularity, reliability, and familiarity. It serves as the relational database for the project, including the double-entry ledger in Gold.

- **RabbitMQ**: For asynchronous communication between services, RabbitMQ was an easy choice due to its simplicity and Docker-friendly setup. Gold consumes `user.created` events from Helium to automatically provision accounts, and publishes transaction events for downstream consumers.

- **Nginx**: Nginx acts as a reverse proxy and load balancer, allowing me to experiment with API gateway patterns and manage service routing efficiently.

- **Docker**: Docker is my go-to tool for containerization, streamlining the orchestration of the entire system and enabling smooth integration of all components.

# Running the Project

Make sure you have installed:

- Docker
- Docker Compose
- .NET SDK
- Node.js

To run the project, first you need to have Docker and Docker Compose installed on your machine. Then, follow these steps:

1. Clone the repository:

```bash
git clone https://github.com/getsnowflake/snowflake
# OR
gh repo clone getsnowflake/snowflake
```

2. Navigate to the project directory:

```bash
cd snowflake
```

3. Start the project:

```bash
make start
```

The required databases and queues will be created, and the services will be started.
