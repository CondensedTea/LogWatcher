LOCAL_BIN ?= ./bin
E2E_LOG_FILE= ./client/short_test.log

version=v1.0.0
container_name=LogWatcher


.PHONY: build
build:
	docker build -t condensedtea/logwatcher:latest -t condensedtea/logwatcher:$(version) .

.PHONY: build-server
build-server:
	go build -o "$(LOCAL_BIN)/LogWatcher" ./server

PHONY: run
run:
	docker run -p 27100:27100/udp --rm -d --name=$(container_name) condensedtea/logwatcher:latest

PHONY: down
down:
	docker kill $(container_name)

.PHONY: client
client:
	go build -o "$(LOCAL_BIN)/TestClient" ./client

PHONY: e2e
e2e:
	./bin/TestClient -log $(E2E_LOG_FILE)