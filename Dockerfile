FROM scratch
EXPOSE 9222
WORKDIR /
COPY domain_exporter .
ENTRYPOINT ["./domain_exporter"]
