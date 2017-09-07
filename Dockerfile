FROM golang:1.9-alpine3.6 AS builder
WORKDIR /go/src/github.com/caarlos0/domain_exporter
ADD . .
RUN apk add -U git
RUN go get -v github.com/golang/dep/...
RUN dep ensure -v
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build --ldflags "-extldflags "-static"" -o domain_exporter .

FROM scratch
EXPOSE 9222
WORKDIR /
COPY --from=builder /go/src/github.com/caarlos0/domain_exporter/domain_exporter .
ENTRYPOINT ["./domain_exporter"]
