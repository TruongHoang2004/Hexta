FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copy the entire workspace
COPY go.work go.work.sum ./
COPY lib ./lib
COPY service ./service

# Build the catalog service
WORKDIR /app/service/catalog
RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service/catalog/main .
COPY --from=builder /app/service/catalog/config ./config

EXPOSE 8083

CMD ["./main"]
