LOCAL_BIN ?= ./bin
E2E_LOG_FILE= ./client/short_test.log

version=v1.0.0

.PHONY: build
build:
	docker build -t condensedtea/logwatcher:latest -t condensedtea/logwatcher:$(version) .

.PHONY: build-server-manually
build-server-manually:
	go build -o "$(LOCAL_BIN)/LogWatcher" ./server

PHONY: run
run:
	docker run -p 27100:27100 --rm -d condensedtea/logwatcher:latest

.PHONY: client
client:
	go build -o "$(LOCAL_BIN)/TestClient" ./client

PHONY: e2e
e2e:
	./bin/TestClient -log $(E2E_LOG_FILE)