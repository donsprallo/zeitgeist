# Builder stage.
FROM golang:1.22-bookworm AS BUILDER

ARG version=0.0.0

WORKDIR /usr/src/app

# Caching module files.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build golang time server daemon.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -o /usr/local/bin/zg-server \
    -ldflags="-X main.version=${version}" \
    ./cmd/zg-server/main.go

# Application image.
FROM alpine:3.19

ARG version=0.0.0

LABEL "org.opencontainers.image.source"="https://github.com/donsprallo/zeitgeist"
LABEL "org.opencontainers.image.version"="${version}"
LABEL "description"="a development NTP time server"

ENV VERSION ${version}

RUN apk add --no-cache \
    ca-certificates tzdata

WORKDIR /usr/src/app

# Copy binary from builder stage.
COPY --from=BUILDER --chown=ntp:ntp \
    /usr/local/bin/zg-server /usr/local/bin/

EXPOSE 123/udp
EXPOSE 80
EXPOSE 443

# Setup time server.
ENV NTP_HOST ""
ENV NTP_PORT 123
ENV WEB_HOST ""
ENV WEB_PORT 80

USER ntp

# Start time server daemon.
CMD ["zg-server"]
