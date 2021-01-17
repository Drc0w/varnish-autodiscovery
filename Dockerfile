FROM golang:alpine AS go-builder

WORKDIR /go/src/github.com/Drc0w/varnish-autodiscovery
COPY . .
RUN apk add --no-cache git && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build

FROM varnish:6.0

COPY --from=go-builder /go/src/github.com/Drc0w/varnish-autodiscovery/varnish-autodiscovery /bin/varnish-autodiscovery
COPY --from=go-builder /go/src/github.com/Drc0w/varnish-autodiscovery/default.tpl ./default.tpl

ENTRYPOINT ["/bin/varnish-autodiscovery"]
