FROM golang:1.24-alpine

WORKDIR ${GOPATH}/avito-pvz/
COPY . ${GOPATH}/avito-pvz/

RUN go build -o /build ./cmd/main.go
RUN go clean -cache -modcache
EXPOSE 8080

CMD ["/build"]