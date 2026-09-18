# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b AS build
ARG TARGETARCH
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /mjolnir ./cmd/mjolnir

FROM alpine:3.24@sha256:5b02b42e375f7426f8d65c3af331ca05d9878f9989230354504e0b9dfd431f60
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
