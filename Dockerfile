FROM golang:alpine

RUN mkdir /opt/proxy

WORKDIR /opt/proxy

COPY . ./

RUN go get \
    && cd run \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go \
    && chmod a+x main

EXPOSE 3128

CMD cd run && ./main
