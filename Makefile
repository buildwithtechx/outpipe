.PHONY: dev dev-api dev-tunnel dev-cron dev-check dev-cli dev-web dev-desktop build fmt format typecheck test clean build-binaries install-cli docker-build docker-images docker-api docker-tunnel docker-cron docker-check

dev:
	npm run dev

dev-api:
	npm run dev:api

dev-tunnel:
	npm run dev:tunnel

dev-cron:
	npm run dev:cron

dev-check:
	npm run dev:check

dev-cli:
	npm run dev:cli

dev-web:
	npm run dev:web

dev-desktop:
	npm run dev:desktop

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
