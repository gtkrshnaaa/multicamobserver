# Stage 1: Build the Go application binary
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy dependency definition files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire workspace code
COPY . .

# Build the monolithic multicam observer binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/multicam-observer ./cmd/web/main.go

# Stage 2: Runtime environment
FROM alpine:3.19

# Install security updates, runtime libraries, and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the compiled binary from Stage 1
COPY --from=builder /app/multicam-observer /app/multicam-observer

# Copy static assets and templates to runtime path
COPY --from=builder /app/ui /app/ui
COPY --from=builder /app/database /app/database

# Expose port 51177 as requested
EXPOSE 51177

# Start the multicam observer application
CMD ["/app/multicam-observer"]
