# Stage 1: Build binary
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy application code
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/main.go

# Stage 2: Final minimal execution environment
FROM alpine:3.19

# Install security certificates for safe external requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy executable from Stage 1
COPY --from=builder /app/main .

# Expose backend API port
EXPOSE 8080

# Run the backend binary
CMD ["./main"]
