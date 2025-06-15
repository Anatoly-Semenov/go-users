FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN apk add --no-cache protoc git

COPY . .

RUN make proto
RUN CGO_ENABLED=0 GOOS=linux go build -o go-users cmd/app/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/go-users /app/go-users
COPY --from=builder /app/migrations /app/migrations

ENV HOST=0.0.0.0 \
    HTTP_PORT=8080 \
    GRPC_PORT=50051 \
    DB_HOST=postgres \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=users \
    DB_SSL_MODE=disable \
    JWT_SECRET_KEY=your-secret-key \
    JWT_EXPIRATION_HOURS=24 \
    LOG_LEVEL=info \
    ENVIRONMENT=production

EXPOSE 8080 50051

ENTRYPOINT ["/app/go-users"]
CMD ["http-server"] 