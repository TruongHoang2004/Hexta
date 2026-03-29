FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copy the entire workspace
COPY go.work go.work.sum ./
COPY lib ./lib
COPY service ./service

# Build the merchant service
WORKDIR /app/service/merchant
RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service/merchant/main .
COPY --from=builder /app/service/merchant/config ./config

EXPOSE 8082

CMD ["./main"]
