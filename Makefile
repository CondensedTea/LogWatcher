LOCAL_BIN ?= ./bin
E2E_LOG_FILE= ./test_client/short_test.log

version=v1.0.0
container_name=LogWatcher

.PHONY: build
build:
	docker build -t condensedtea/logwatcher:latest -t condensedtea/logwatcher:$(version) .

.PHONY: build-local
build-local:
	go build -o "$(LOCAL_BIN)/LogWatcher" ./server
	go build -o "$(LOCAL_BIN)/TestClient" ./test_client

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
	go test -race .