FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git

# Copy workspace files
COPY go.work go.work.sum* ./
COPY packages/shared/go.mod packages/shared/go.sum* ./packages/shared/
COPY services/api/go.mod services/api/go.sum* ./services/api/
COPY services/connector/go.mod services/connector/go.sum* ./services/connector/
COPY test/go.mod test/go.sum* ./test/

RUN go mod download

COPY packages/shared ./packages/shared
COPY services/connector ./services/connector

WORKDIR /app/services/connector
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/connector ./cmd/connector

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 connector

COPY --from=builder /app/bin/connector .

USER connector

EXPOSE 8091

CMD ["./connector"]
