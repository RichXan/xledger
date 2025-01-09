FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api-gateway ./api-gateway/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/user-service ./user-service/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/post-service ./post-service/main.go

# Final stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/api-gateway /app/bin/api-gateway
COPY --from=builder /app/bin/user-service /app/bin/user-service
COPY --from=builder /app/bin/post-service /app/bin/post-service

# Copy configuration files if needed
COPY config/ /app/config/

# Expose necessary ports (adjust according to your needs)
EXPOSE 8080 8081 8082

# Command to run the application
CMD ["/app/bin/api-gateway"]
