FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o taxapp ./cmd/taxapp

# Use a small alpine image
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the config files
COPY --from=builder /app/config/config.yaml /root/config/config.yaml
COPY --from=builder /app/config/config.prod.yaml /root/config/config.prod.yaml

# Copy the binary from builder
COPY --from=builder /app/taxapp /root/

# Command to run when the container starts
ENTRYPOINT ["./taxapp"]