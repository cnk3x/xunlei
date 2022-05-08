push:
	docker buildx build -t cnk3x/xunlei:latest -t cnk3x/xunlei:2.7.1 --platform=linux/amd64,linux/arm64 --push .

build:
	docker buildx build -t cnk3x/xunlei:latest -t cnk3x/xunlei:2.7.1 --platform=linux/amd64 --load .

up: 
	docker compose up -d

down:
	docker compose down

clean: down
	rm -rf data

log:
	docker compose logs -f

all: down build run log
