FROM golang:1.26-alpine AS builder

ARG VERSION

WORKDIR /usr/app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

ENV CGO_ENABLED=0
RUN go build -buildvcs=false -o bin/service -ldflags="-X main.Version=${VERSION}" ./cmd/service

FROM alpine:3.24.1

ARG VERSION

LABEL org.opencontainers.image.title="audited" \
      org.opencontainers.image.description="Audit log management service for cloud-native applications" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.source="https://github.com/getaudited/audited"

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S audited && adduser -S audited -G audited

WORKDIR /usr/app
RUN mkdir -p misc/sql/migrations
COPY --from=builder /usr/app/misc/sql/migrations/* ./misc/sql/migrations/
COPY --from=builder /usr/app/bin/service ./service

USER audited

ENTRYPOINT ["./service"]
