VERSION := 3.0.2

version:
	@echo $(VERSION)

localhost:
	docker buildx build -t localhost/w7x/xunlei:$(VERSION) --load .

push:
	docker buildx build -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .

ghcr:
	docker buildx build -t ghcr.io/cnk3x/xunlei:$(VERSION) -t ghcr.io/cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .
