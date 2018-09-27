.PHONY: build run-dev build-docker

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
DEV_IMAGE := tent-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))
DEV_DOCKER_IMAGE := tent-bin-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))

default: binary

binary: build
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=linux -e CGO_ENABLED=0 -e GOGC=off -e GOARCH=amd64 "$(DEV_IMAGE)" go build -a -tags netgo -ldflags '-w' -o "/dist/tent"

crossbinary: binary
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=linux -e GOARCH=amd64 "$(DEV_IMAGE)" go build -o "/dist/tent-linux-amd64"
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=linux -e GOARCH=386 "$(DEV_IMAGE)" go build -o "/dist/tent-linux-386"
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=darwin -e GOARCH=amd64 "$(DEV_IMAGE)" go build -o "/dist/tent-darwin-amd64"
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=darwin -e GOARCH=386 "$(DEV_IMAGE)" go build -o "/dist/tent-darwin-386"
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=windows -e GOARCH=amd64 "$(DEV_IMAGE)" go build -o "/dist/tent-windows-amd64.exe"
	docker run -it -v $(CURDIR)/dist:/dist -e GO111MODULE=on -e GOOS=windows -e GOARCH=386 "$(DEV_IMAGE)" go build -o "/dist/tent-windows-386.exe"

install:
	go generate

test: build
	docker run -it -e GO111MODULE=on "$(DEV_IMAGE)" go test "./..."

coverage: build
	docker run -it -e GO111MODULE=on "$(DEV_IMAGE)" ./script/coverage.sh

build: install dist
	docker build -t "$(DEV_IMAGE)" --target build -f build.Dockerfile .

dist:
	mkdir dist

run-dev:
	go generate
	go test ./...
	go build -o "tent"
	./tent

build-docker:
	docker build -t "$(DEV_DOCKER_IMAGE)" .
