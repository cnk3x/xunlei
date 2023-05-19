VERSION := 3.7.1
HOMER := $(shell cat home.repo.txt)
GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

version::
	@echo $(VERSION)

localhost::
	docker buildx build --load -t local/xunlei:$(VERSION) .

home::
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(HOMER)/xunlei:$(VERSION) -t $(HOMER)/xunlei:latest \
	.

dockerhub::
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(HUB)/xunlei:$(VERSION)   -t $(HUB)/xunlei:latest \
	.

ghcr::
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:$(VERSION)   -t $(GHR)/xunlei:latest \
	.

aliyun::
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(ALIR)/xunlei:$(VERSION)  -t $(ALIR)/xunlei:latest \
	.

all::
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:$(VERSION)   -t $(GHR)/xunlei:latest \
	-t $(ALIR)/xunlei:$(VERSION)  -t $(ALIR)/xunlei:latest \
	-t $(HUB)/xunlei:$(VERSION)   -t $(HUB)/xunlei:latest \
	.
