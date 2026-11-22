# Mock provider platform and load generator for local runs; never deployed to production.
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

COPY test ./test

WORKDIR /app/test
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/mockplatform ./cmd/mockplatform \
    && CGO_ENABLED=0 GOOS=linux go build -o /app/bin/loadgen ./cmd/loadgen

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 mock

COPY --from=builder /app/bin/mockplatform /app/bin/loadgen ./

USER mock

EXPOSE 8090

CMD ["./mockplatform"]
