FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copy the entire workspace
COPY go.work go.work.sum ./
COPY lib ./lib
COPY service ./service

# Build the api service
WORKDIR /app/service/api
RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service/api/main .
COPY --from=builder /app/service/api/config ./config

EXPOSE 9090

CMD ["./main"]
