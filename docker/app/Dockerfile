FROM golang:1.18.1-alpine3.15

RUN apk add --no-cache git

RUN GOPATH="/go/pkg/mod"

# コンテナの作業ディレクトリを変更
RUN mkdir /go/src/app
COPY ./server/go.mod ./server/go.sum /go/src/app/
WORKDIR /go/src/app
# ADD ./ /go/src/app
RUN go mod download
