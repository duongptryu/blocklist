FROM golang AS builder

RUN git clone --branch v1.5.0 --depth 1 https://github.com/coredns/coredns.git

WORKDIR coredns

COPY plugin.cfg .

RUN go generate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o  /go/bin/coredns

FROM scratch

COPY --from=builder /go/bin/coredns /coredns
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY Corefile /etc/coredns/Corefile

EXPOSE 53/udp
EXPOSE 9153/tcp

ENTRYPOINT ["/coredns", "-conf", "/etc/coredns/Corefile"]