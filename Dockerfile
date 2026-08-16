FROM alpine:3.22

LABEL org.opencontainers.image.title="mjolnir" \
      org.opencontainers.image.description="Modern NUT UPS monitoring container" \
      org.opencontainers.image.source="https://github.com/st0o0/mjolnir"

RUN apk add --no-cache \
      nut \
      tini \
      libusb \
      net-snmp-libs \
    && mkdir -p /run/nut /etc/nut/local \
    && chown -R nut:nut /run/nut

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENV NUT_UPS_1_NAME=ups \
    NUT_UPS_1_DRIVER=usbhid-ups \
    NUT_UPS_1_PORT=auto \
    NUT_UPS_1_DESC=UPS \
    NUT_USER=admin \
    NUT_PASSWORD= \
    NUT_SECRET_NAME=nut-password \
    NUT_SERVER=primary \
    NUT_LISTEN=0.0.0.0 \
    NUT_MAXAGE=15

EXPOSE 3493

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD upsc ${NUT_UPS_1_NAME}@localhost:3493 ups.status 2>/dev/null || exit 1

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/entrypoint.sh"]
