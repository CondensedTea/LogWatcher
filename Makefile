LOCAL_BIN ?= ./bin

version=v1.0.0
container_name=LogWatcher

.DEFAULT_GOAL := default

.PHONY: default
default: build run

.PHONY: build
build:
	docker build -t condensedtea/logwatcher:latest -t condensedtea/logwatcher:$(version) .

.PHONY: build-local
build-local:
	CGO_ENABLED=0 go build -o "$(LOCAL_BIN)/LogWatcher" ./server
	CGO_ENABLED=0 go build -o "$(LOCAL_BIN)/TestClient" ./e2e

PHONY: run
run:
	docker run --network=host --rm -d --name=$(container_name) condensedtea/logwatcher:latest

PHONY: down
down:
	docker kill $(container_name)

PHONY: e2e
e2e:
	./bin/TestClient -log $(E2E_LOG_FILE)

PHONY: test
test:
	go test -race ./server