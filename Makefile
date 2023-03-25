VERSION := 3.5.2
HOMEREPO := $(shell cat home.repo.txt)
GHREPO := ghcr.io/cnk3x
ALIREPO := registry.cn-shenzhen.aliyuncs.com/cnk3x

version:
	@echo $(VERSION)

localhost:
	docker buildx build --load -t localhost/xunlei:$(VERSION) .

home:
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(HOMEREPO)/xunlei:$(VERSION)  -t $(HOMEREPO)/xunlei:latest \
	.

push:
	docker buildx build -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .

ghcr:
	docker buildx build -t $(GHREPO)/xunlei:$(VERSION) -t $(GHREPO)/xunlei:latest --platform linux/amd64,linux/arm64 --push .

aliyun:
	docker buildx build -t $(ALIREPO)/xunlei:$(VERSION) -t $(ALIREPO)/xunlei:latest --platform linux/amd64,linux/arm64 --push .

all::
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHREPO)/xunlei:$(VERSION)  -t $(GHREPO)/xunlei:latest \
	-t $(ALIREPO)/xunlei:$(VERSION) -t $(ALIREPO)/xunlei:latest \
	-t cnk3x/xunlei:$(VERSION)      -t cnk3x/xunlei:latest
