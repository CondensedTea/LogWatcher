LOCAL_BIN ?= ./bin

version=v1.1.4
container_name=LogWatcher
loglevel=info
config_path=config.yaml

LDFLAGS = "-X 'LogWatcher/pkg/requests.Version=$(version)' -X 'main.LogLevel=$(loglevel)' -X 'main.ConfigPath=$(config_path)'"

.DEFAULT_GOAL := default

.PHONY: default
default: build run

.PHONY: build
build:
	docker build -t condensedtea/logwatcher:latest -t condensedtea/logwatcher:$(version) .

.PHONY: build-local
	app e2e

.PHONY: app
app:
	CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o "$(LOCAL_BIN)/LogWatcher" ./app

.PHONY: e2e
e2e:
	CGO_ENABLED=0 go build -o "$(LOCAL_BIN)/TestClient" ./e2e

PHONY: run
run:
	docker run --network=host --rm -d --name=$(container_name) condensedtea/logwatcher:latest

PHONY: down
down:
	docker kill $(container_name)

PHONY: e2e-test
e2e-test:
	./bin/TestClient -log $(E2E_LOG_FILE)

PHONY: test
test:
	go test -race ./...
