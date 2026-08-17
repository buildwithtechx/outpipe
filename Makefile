.PHONY: dev build fmt format typecheck test clean build-binaries install-cli docker-build docker-images docker-api docker-tunnel docker-cron docker-check

dev:
	npm run dev

build:
	npm run build

fmt:
	npm run fmt

format:
	npm run format

typecheck:
	npm run typecheck

test:
	npm run test

clean:
	npm run clean

build-binaries:
	bash scripts/build-binaries.sh

install-cli:
	bash scripts/install-cli.sh

docker-build:
	docker build -f docker/Dockerfile.api -t outpipe-api:dev .

docker-images:
	bash scripts/build-images.sh

docker-api:
	docker build -f docker/Dockerfile.api -t outpipe-api:dev .

docker-tunnel:
	docker build -f docker/Dockerfile.tunnel -t outpipe-server:dev .

docker-cron:
	docker build -f docker/Dockerfile.cron -t outpipe-cron:dev .

docker-check:
	docker build -f docker/Dockerfile.check -t outpipe-check:dev .
