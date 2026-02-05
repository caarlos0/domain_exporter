FROM alpine:3.21
RUN apk upgrade --no-cache
EXPOSE 9222
ENTRYPOINT ["/usr/bin/domain_exporter"]
COPY domain_exporter_*.apk /tmp/
RUN apk add --allow-untrusted /tmp/domain_exporter_*.apk
