LOCAL_BIN ?= ./bin
E2E_LOG_FILE= ./client/short_test.log

.PHONY: build
build: build-server build-client

.PHONY: build-server
build-server:
	go build -o "$(LOCAL_BIN)/LogWatcher" ./server

.PHONY: build-client
build-client:
	go build -o "$(LOCAL_BIN)/TestClient" ./client

PHONY: run
run:
	./bin/LogWatcher

PHONY: e2e
e2e:
	./bin/TestClient -log $(E2E_LOG_FILE)