FROM alpine
MAINTAINER Anthony Woods awoods@raintank.io

RUN apk --update add ca-certificates

RUN mkdir -p /etc/raintank
COPY config/probe.ini /etc/raintank/probe.ini
COPY config/publicChecks.json /etc/raintank/publicChecks.json

COPY build/raintank-probe /usr/bin/raintank-probe
COPY entrypoint.sh /usr/bin/

ENTRYPOINT ["/usr/bin/entrypoint.sh"]
CMD ["-config=/etc/raintank/probe.ini"]
