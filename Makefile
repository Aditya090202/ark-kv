.PHONY: build run test clean

BINARY=miniKV
PORT=6379

build:
	go build -o $(BINARY) ./cmd/server/

run: build
	./$(BINARY) --port $(PORT)

run-redis:
	./$(BINARY) --port 6379

test:
	go test -v -race ./...

test-short:
	go test -short -v ./...

bench:
	go test -bench=. ./...

lint:
	test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

clean:
	rm -f $(BINARY)
	go clean

# Interactive: connect with redis-cli
cli:
	redis-cli -p $(PORT)

# Quick smoke test: SET then GET
smoke:
	@echo "SET test hello / GET test"
	@printf "*3\r\n\$$3\r\nSET\r\n\$$4\r\ntest\r\n\$$5\r\nhello\r\n" | nc -q1 localhost $(PORT)
	@printf "*2\r\n\$$3\r\nGET\r\n\$$4\r\ntest\r\n" | nc -q1 localhost $(PORT)

# Benchmark: flood SET commands
bench-flood:
	@echo "Sending 1000 SET commands via redis-benchmark (if available)"
	redis-benchmark -p $(PORT) -n 1000 -t SET,GET -q 2>/dev/null || echo "redis-benchmark not installed; use: redis-benchmark -p $(PORT) -n 1000 -t SET,GET"

docker-build:
	docker build -t miniKV .

docker-run:
	docker run --rm -p $(PORT):6379 miniKV --port 6379