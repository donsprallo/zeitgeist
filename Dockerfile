# Builder stage
FROM golang:1.19.3-bullseye AS BUILDER

WORKDIR /usr/src/app

# Caching module files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build golang time server daemon.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -o /usr/local/bin/gotsd \
    ./cmd/gotsd/main.go

# Application image
FROM golang:1.19.3-alpine3.16

RUN apk add --no-cache \
    ca-certificates bash

WORKDIR /usr/src/app

# Copy binary from builder stage
COPY --from=BUILDER --chown=ntp:ntp \
    /usr/local/bin/gotsd /usr/local/bin/

EXPOSE 123/udp
EXPOSE 80
EXPOSE 443

# Setup time server
ENV NTP_HOST ""
ENV NTP_PORT 123
ENV API_HOST ""
ENV API_PORT 80

USER ntp

# Start time server daemon
CMD ["gotsd"]
