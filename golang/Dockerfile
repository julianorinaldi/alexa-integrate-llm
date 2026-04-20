FROM golang:1.24-alpine AS builder

WORKDIR /app

# Initialize dependencies
COPY . .
RUN go mod tidy

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# ---- Final Stage ----
FROM alpine:3.18

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/server .

# Run the binary
CMD ["./server"]
