FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy workspace files
COPY go.work go.work.sum ./
COPY packages/shared/go.mod packages/shared/go.sum* ./packages/shared/
COPY services/api/go.mod services/api/go.sum* ./services/api/

# Download dependencies
RUN go mod download

# Copy source code
COPY packages/shared ./packages/shared
COPY services/api ./services/api

# Build the api service
WORKDIR /app/services/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/api .
COPY --from=builder /app/services/api/config ./config

EXPOSE 8080

CMD ["./api"]
