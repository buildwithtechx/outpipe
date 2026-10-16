.PHONY: all build build-api build-tunnel build-cron build-check build-cli test fmt tidy clean docker-api docker-tunnel docker-cron docker-check

all: build

build: build-api build-tunnel build-cron build-check build-cli

build-api:
	go build -ldflags="-s -w" -o bin/outpipe-api ./cmd/server

build-tunnel:
	go build -ldflags="-s -w" -o bin/outpipe-server ./cmd/tunnel

build-cron:
	go build -ldflags="-s -w" -o bin/tunnel-cron ./cmd/cron

build-check:
	go build -ldflags="-s -w" -o bin/tunnel-check ./cmd/check

build-cli:
	go build -ldflags="-s -w" -o bin/outpipe ./cmd/cli

test:
	go test -count=1 ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy
	go mod vendor

clean:
	rm -rf bin

docker-api:
	docker build -f docker/Dockerfile.api -t outpipe-api:latest .

docker-tunnel:
	docker build -f docker/Dockerfile.tunnel -t outpipe-tunnel:latest .

docker-cron:
	docker build -f docker/Dockerfile.cron -t outpipe-cron:latest .

docker-check:
	docker build -f docker/Dockerfile.check -t outpipe-check:latest .
