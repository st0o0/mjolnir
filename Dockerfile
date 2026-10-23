# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS build
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /mjolnir ./cmd/mjolnir

FROM alpine:3.24@sha256:71e9b7bae160ba167fa17414401b94cbcdaf3c32feeab9298bd794a2a0614c35
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
