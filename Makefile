.PHONY: all build release build-multiarch build-multiarch-push setup-buildx

IMAGE=dddpaul/finparser

all: build

build-alpine:
	mkdir -p ./bin
	CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -ldflags="-w -s" -o ./bin/finparser ./finparser.go

build:
	@docker build --tag=${IMAGE} .

build-multiarch:
	@docker buildx build --platform linux/amd64,linux/arm64 --tag=${IMAGE} .

build-multiarch-push:
	@docker buildx build --platform linux/amd64,linux/arm64 --tag=${IMAGE} --push .

setup-buildx:
	@echo "Setting up Docker Buildx for multiarch builds..."
	@docker buildx create --name multiarch --driver docker-container --use || true
	@docker buildx inspect --bootstrap

debug:
	@docker run -it --entrypoint=sh ${IMAGE}

release: build
	@echo "Tag image with version $(version)"
	@docker tag ${IMAGE} ${IMAGE}:$(version)

release-multiarch:
	@echo "Building and tagging multiarch image with version $(version)"
	@docker buildx build --platform linux/amd64,linux/arm64 --tag=${IMAGE} --tag=${IMAGE}:$(version) --push .

push: release
	@docker push ${IMAGE}
	@docker push ${IMAGE}:$(version)

push-multiarch: release-multiarch
	@echo "Multiarch images pushed successfully"
