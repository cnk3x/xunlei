VERSION:=2.8.0

build:
	docker buildx build -t cnk3x/xunlei:latest --load .

push:
	docker buildx build -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform=linux/amd64,linux/arm64 --push .

up: 
	docker compose up -d

down:
	docker compose down

clean: down
	rm -rf data
	rm -rf tmp
	rm -rf local/xunlei

log:
	docker compose logs -f

all: down build
	docker compose up
