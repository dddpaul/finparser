.PHONY: all build release

IMAGE=dddpaul/finparser
VERSION=1.3

all: build

build-alpine:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/finparser ./finparser.go

build:
	@docker build --tag=${IMAGE} .

debug:
	@docker run -it --entrypoint=sh ${IMAGE}

release: build
	@docker build --tag=${IMAGE}:${VERSION} .

deploy: release
	@docker push ${IMAGE}
	@docker push ${IMAGE}:${VERSION}
