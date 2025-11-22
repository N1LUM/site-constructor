# Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -o main ./main.go

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/.env .env

CMD ["./main"]
