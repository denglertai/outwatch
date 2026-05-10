BINARY ?= outwatch
CMD_PATH ?= ./cmd/outwatch
DIST_DIR ?= ./dist
IMAGE ?= outwatch
TAG ?= latest
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: help test build clean image-build image-build-no-cache image-buildx-push test-docker-setup test-docker-up test-docker-down

help:
	@echo "Targets:"
	@echo "  test                   Run Go tests"
	@echo "  build                  Build CLI binary into $(DIST_DIR)/$(BINARY)"
	@echo "  clean                  Remove build artifacts"
	@echo "  image-build            Build container image $(IMAGE):$(TAG)"
	@echo "  image-build-no-cache   Build container image without cache"
	@echo "  image-buildx-push      Build and push multi-arch image via buildx"
	@echo "  test-docker-setup      Create ./generated directory with correct permissions"
	@echo "  test-docker-up         Start Docker Compose example with correct setup"
	@echo "  test-docker-down       Stop Docker Compose example"
	@echo ""
	@echo "Parameters (override with VAR=value):"
	@echo "  BINARY=$(BINARY)       Output binary name for build target"
	@echo "  CMD_PATH=$(CMD_PATH)   Go package path used as build entrypoint"
	@echo "  DIST_DIR=$(DIST_DIR)   Output directory for build artifacts"
	@echo "  IMAGE=$(IMAGE)         Container image repository/name"
	@echo "  TAG=$(TAG)             Container image tag"
	@echo "  PLATFORMS=$(PLATFORMS) Target platforms for buildx multi-arch builds"

test:
	go test -coverprofile=coverage.txt ./...

build:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -o $(DIST_DIR)/$(BINARY) $(CMD_PATH)

clean:
	rm -rf $(DIST_DIR)

image-build:
	docker build -t $(IMAGE):$(TAG) .

image-build-no-cache:
	docker build --no-cache -t $(IMAGE):$(TAG) .

image-buildx-push:
	@docker buildx inspect outwatch-builder >/dev/null 2>&1 || docker buildx create --name outwatch-builder --use
	@docker buildx use outwatch-builder
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE):$(TAG) --push .

test-docker-setup:
	mkdir -p ./examples/docker/generated && chmod 777 ./examples/docker/generated

test-docker-up: test-docker-setup
	cd ./examples/docker && docker compose up --build

test-docker-down:
	cd ./examples/docker && docker compose down
