FROM golang:1.15-alpine as builder

WORKDIR /usr/local/go/src/gomq

ADD . /usr/local/go/src/gomq

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add git && git config --global https.proxy http://127.0.0.1:8118 &&  git config --global https.proxy https://127.0.0.1:8118

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o gomq server/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o gomqctl cmd/main.go

FROM alpine as runner

COPY --from=builder /usr/local/go/src/gomq/gomq     /usr/local/bin
COPY --from=builder /usr/local/go/src/gomq/gomqctl  /usr/local/bin

CMD gomq