# Builder stage
FROM golang:1.19.4-bullseye AS BUILDER

ARG version=0.0.0
ARG hash=---

WORKDIR /usr/src/app

# Caching module files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build golang time server daemon.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -o /usr/local/bin/gotsd \
    -ldflags="-X main.version=${version} \
      -X main.hash=${hash}" \
    ./cmd/gotsd/main.go

# Application image
FROM golang:1.19.4-alpine3.17

ARG version=0.0.0
ARG hash=---

LABEL "org.opencontainers.image.source"="https://github.com/donsprallo/gots"
LABEL "version"="${version}"
LABEL "hash"="${hash}"
LABEL "description"="a development NTP time server"

ENV VERSION ${version}
ENV HASH ${hash}

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
ENV WEB_HOST ""
ENV WEB_PORT 80

USER ntp

# Start time server daemon
CMD ["gotsd"]
