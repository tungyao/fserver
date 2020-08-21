FROM golang:1.14-alpine
WORKDIR /go/src/app
COPY . .
RUN mkdir -p /go/src/app/log
RUN mkdir -p /go/src/app/mount
RUN mkdir -p /go/src/app/quelity
RUN touch /go/src/app/log/fserver.log
RUN go build fserver.go
RUN ls
EXPOSE 8105:8105

ARG domino=localhost\/
ENV domino $domino
ARG user=admin
ENV user $user
ARG pass=admin
ENV pass $user

CMD /go/src/app/fserver -domino=${domino} -user=${user} -pass=${pass}