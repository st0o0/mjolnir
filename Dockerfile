# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /mjolnir ./cmd/mjolnir

FROM alpine:3.22 AS runtime
LABEL org.opencontainers.image.title="mjolnir" \
      org.opencontainers.image.description="Modern NUT UPS monitoring container with Prometheus metrics" \
      org.opencontainers.image.source="https://github.com/st0o0/mjolnir" \
      org.opencontainers.image.documentation="https://github.com/st0o0/mjolnir#readme" \
      org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache \
      nut \
      tini \
      libusb \
      net-snmp-libs \
    && mkdir -p /run/nut /etc/nut/local \
    && chown -R nut:nut /run/nut

COPY --from=build /mjolnir /usr/local/bin/mjolnir

ENV NUT_UPS_1_NAME=ups \
    NUT_UPS_1_DRIVER=usbhid-ups \
    NUT_UPS_1_PORT=auto \
    NUT_UPS_1_DESC=UPS \
    NUT_USER=admin \
    NUT_SECRET_NAME=nut-password \
    NUT_SERVER=primary \
    NUT_LISTEN=0.0.0.0 \
    NUT_MAXAGE=15

EXPOSE 3493 9550

HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 \
  CMD ["/usr/local/bin/mjolnir", "healthcheck"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/mjolnir"]
