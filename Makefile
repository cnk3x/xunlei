VERSION := 3.1.8
HOMEREPO := $(cat code.home.casaos.cn/w7x)

version:
	@echo $(VERSION)

localhost:
	docker buildx build -t localhost/xunlei:$(VERSION) --load .

home:
	docker buildx build -t $(HOMEREPO)/xunlei:$(VERSION) --push .

push:
	docker buildx build -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .

ghcr:
	docker buildx build -t ghcr.io/cnk3x/xunlei:$(VERSION) -t ghcr.io/cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .
