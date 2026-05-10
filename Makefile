BINARY ?= outwatch
CMD_PATH ?= ./cmd/outwatch
DIST_DIR ?= ./dist
IMAGE ?= outwatch
TAG ?= latest

.PHONY: help test build clean image-build image-build-no-cache test-docker-setup test-docker-up test-docker-down

help:
	@echo "Targets:"
	@echo "  test                   Run Go tests"
	@echo "  build                  Build CLI binary into $(DIST_DIR)/$(BINARY)"
	@echo "  clean                  Remove build artifacts"
	@echo "  image-build            Build container image $(IMAGE):$(TAG)"
	@echo "  image-build-no-cache   Build container image without cache"
	@echo "  test-docker-setup      Create ./generated directory with correct permissions"
	@echo "  test-docker-up         Start Docker Compose example with correct setup"
	@echo "  test-docker-down       Stop Docker Compose example"

test:
	go test ./...

build:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -o $(DIST_DIR)/$(BINARY) $(CMD_PATH)

clean:
	rm -rf $(DIST_DIR)

image-build:
	docker build -t $(IMAGE):$(TAG) .

image-build-no-cache:
	docker build --no-cache -t $(IMAGE):$(TAG) .

test-docker-setup:
	mkdir -p ./examples/docker/generated && chmod 777 ./examples/docker/generated

test-docker-up: test-docker-setup
	cd ./examples/docker && docker compose up --build

test-docker-down:
	cd ./examples/docker && docker compose down
