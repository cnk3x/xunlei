VERSION := 3.1.10
HOMEREPO := $(shell cat home.repo.txt)

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

aliyun:
	docker buildx build -t registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:$(VERSION) -t registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .
