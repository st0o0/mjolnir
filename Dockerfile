# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /mjolnir ./cmd/mjolnir

FROM alpine:3.24@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
LABEL org.opencontainers.image.title="mjolnir" \
      org.opencontainers.image.description="Modern NUT UPS monitoring container with Prometheus metrics" \
      org.opencontainers.image.source="https://github.com/st0o0/mjolnir" \
      org.opencontainers.image.documentation="https://github.com/st0o0/mjolnir#readme" \
      org.opencontainers.image.licenses="MIT"

RUN apk upgrade --no-cache \
    && apk add --no-cache \
      nut \
      tini \
      libusb \
      net-snmp-libs \
    && mkdir -p /run/nut /etc/nut/local \
    && chown -R nut:nut /run/nut

COPY --from=build /mjolnir /usr/local/bin/mjolnir
COPY LICENSE /

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
