# Stage 1: Build the application using Maven
FROM maven:3.9.8-eclipse-temurin-22 AS build

# Set the working directory in the container
WORKDIR /app

# Install the protobuf compiler in a single RUN command
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends --assume-yes \
      protobuf-compiler && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copy the pom.xml file and download dependencies
COPY pom.xml .
RUN mvn dependency:go-offline

# Copy the rest of your application
COPY src ./src

# Build the application with verbose logging
RUN mvn -X package -DskipTests

# Stage 2: Use a smaller image for the final stage
FROM eclipse-temurin:22-jre-alpine

# Set the working directory in the container
WORKDIR /app

# Copy the built artifact from the build stage
COPY --from=build /app/target/scorer-0.0.1-SNAPSHOT.jar ./app.jar

# Expose the port your application runs on
EXPOSE 8080

# Set Spring profile to prod for Docker environment
ENV SPRING_PROFILES_ACTIVE=prod

# Specify the command to run on container start
CMD ["java", "-jar", "app.jar"]
