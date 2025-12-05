# Build stage
FROM golang:latest AS builder

WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary (main is under cmd/vital-sync)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/vital-sync

# Final stage for Go app
FROM alpine:3.20 AS final

WORKDIR /app

# Update CA certificates and tzdata for time operations
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/main /app/main
COPY --from=builder /app/config /app/config

EXPOSE 8080

ENTRYPOINT ["/app/main"]
