FROM golang:1.11-alpine as dev

RUN apk --no-cache --no-progress add git bash gcc musl-dev make curl tar \
    && mkdir -p /go/src/labs.pmsystem.co.uk/devops/tent

RUN go get -u golang.org/x/lint/golint \
    && go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/labs.pmsystem.co.uk/devops/tent

FROM dev as build

COPY . /go/src/labs.pmsystem.co.uk/devops/tent