# Builder stage
FROM golang:1.19.3-bullseye AS BUILDER

WORKDIR /usr/src/app

# Cache
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
# Build golang time server daemon.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -v -o /usr/local/bin/gotsd \
    ./cmd/ntp-server/main.go

# Application image
FROM golang:1.19.3-alpine3.16

RUN apk add --no-cache \
    ca-certificates bash

WORKDIR /usr/src/app

COPY --from=BUILDER --chown=ntp:ntp \
    /usr/local/bin/gotsd /usr/local/bin/

EXPOSE 123
EXPOSE 80
EXPOSE 443

ENV NTP_HOST "0.0.0.0"
ENV NTP_PORT 123

USER ntp
CMD ["gotsd"]