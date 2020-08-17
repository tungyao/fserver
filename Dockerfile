FROM golang:1.14-alpine
WORKDIR /go/src/app
COPY . .
RUN mkdir -p /go/src/app/log
RUN mkdir -p /go/src/app/mount
RUN mkdir -p /go/src/app/quelity
RUN touch /go/src/app/log/fserver.log
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add git
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]