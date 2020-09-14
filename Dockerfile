FROM golang:1.14-alpine as builder
RUN apk add --no-cache \
    git
WORKDIR /opt/cfdnsproxy

COPY . .
RUN go get
RUN go build

FROM alpine as service
RUN apk add --no-cache \
    bind-tools
COPY --from=builder /opt/cfdnsproxy/cfdnsproxy /usr/bin/cfdnsproxy
RUN chmod +x /usr/bin/cfdnsproxy
EXPOSE 2000/tcp
HEALTHCHECK --interval=10s --timeout=3s \
  CMD  dig +timeout=2 +tcp example.com. @localhost -p 2000 || exit 1
CMD ["/usr/bin/cfdnsproxy"]


