FROM golang:1.11.4-alpine3.8 as builder

ENV GOPATH /go

RUN apk add make

WORKDIR /go/src/GoExcercise

ADD . .

RUN make build-server

FROM alpine:3.8

EXPOSE 8080

RUN apk add openssl ca-certificates

COPY --from=builder /go/src/GoExcercise/bin/server /tmp/

RUN chmod +x /tmp/server

ENTRYPOINT /tmp/server