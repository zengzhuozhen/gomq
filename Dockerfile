FROM golang:1.15-alpine as builder

WORKDIR /usr/local/go/src/gomq

ADD . /usr/local/go/src/gomq

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o gomq server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o gomqctl cmd/main.go

FROM alpine as runner

COPY --from=builder /usr/local/go/src/gomq/gomq     /usr/local/bin
COPY --from=builder /usr/local/go/src/gomq/gomqctl  /usr/local/bin

CMD gomq