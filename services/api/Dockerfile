FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install build dependencies
RUN apk add --no-cache git

# Copy workspace files
COPY go.work go.work.sum* ./
COPY packages/shared/go.mod packages/shared/go.sum* ./packages/shared/
COPY services/api/go.mod services/api/go.sum* ./services/api/
COPY services/connector/go.mod services/connector/go.sum* ./services/connector/
COPY test/go.mod test/go.sum* ./test/

# Download dependencies
RUN go mod download

# Copy source code
COPY packages/shared ./packages/shared
COPY services/api ./services/api

# Build the api service
WORKDIR /app/services/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/platform-operator ./cmd/platform-operator
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/ingress ./cmd/ingress
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/messaging-command ./cmd/messaging-command

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/api .
COPY --from=builder /app/bin/platform-operator .
# Kafka consumer for ingress, outbound results and connector health; run with command ./ingress.
COPY --from=builder /app/bin/ingress .
# Conversation write side; the public API forwards claim/assign/handoff/reply routes here. Run with ./messaging-command.
COPY --from=builder /app/bin/messaging-command .
COPY --from=builder /app/services/api/config ./config

EXPOSE 8080

CMD ["./api"]
