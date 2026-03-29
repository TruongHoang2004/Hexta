FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copy the entire workspace
COPY go.work go.work.sum ./
COPY lib ./lib
COPY service ./service

# Build the user service
WORKDIR /app/service/user
RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service/user/main .
COPY --from=builder /app/service/user/config ./config

EXPOSE 8080

CMD ["./main"]
