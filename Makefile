VERSION := 3.4.0
HOMEREPO := $(shell cat home.repo.txt)
GHREPO := ghcr.io/cnk3x
ALIREPO := registry.cn-shenzhen.aliyuncs.com/cnk3x

# 临时更新版本
CLI_UPDATE_VERSION := 3.4.0
CLI_UPDATE_VERSION_DATE := 20230209
CLI_UPDATE_URL_PREFIX := https://static.xbase.cloud/file/2rvk4e3gkdnl7u1kl0k/pancli

version:
	@echo $(VERSION)

localhost:
	docker buildx build -t localhost/xunlei:$(VERSION) --load .

home:
	docker buildx build -t $(HOMEREPO)/xunlei:$(VERSION) --push .

push:
	docker buildx build -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .

ghcr:
	docker buildx build -t $(GHREPO)/xunlei:$(VERSION) -t $(GHREPO)/xunlei:latest --platform linux/amd64,linux/arm64 --push .

aliyun:
	docker buildx build -t $(ALIREPO)/xunlei:$(VERSION) -t $(ALIREPO)/xunlei:latest --platform linux/amd64,linux/arm64 --push .

push:
	docker buildx build \
	-t $(HOMEREPO)/xunlei:$(VERSION) \
	-t $(HOMEREPO)/xunlei:latest \
	-t $(GHREPO)/xunlei:$(VERSION) \
	-t $(GHREPO)/xunlei:latest \
	-t $(ALIREPO)/xunlei:$(VERSION) \
	-t $(ALIREPO)/xunlei:latest \
	-t cnk3x/xunlei:$(VERSION) \
	-t cnk3x/xunlei:latest \
	--platform linux/amd64,linux/arm64 \
	--push .

update:
	docker buildx build \
	--build-arg CLI_UPDATE_VERSION=$(CLI_UPDATE_VERSION) \
    --build-arg CLI_UPDATE_VERSION_DATE=$(CLI_UPDATE_VERSION_DATE) \
    --build-arg CLI_UPDATE_URL_PREFIX=$(CLI_UPDATE_URL_PREFIX) \
	-t $(HOMEREPO)/xunlei:$(CLI_UPDATE_VERSION) \
	-t $(GHREPO)/xunlei:$(CLI_UPDATE_VERSION) \
	-t $(ALIREPO)/xunlei:$(CLI_UPDATE_VERSION) \
	-t cnk3x/xunlei:$(CLI_UPDATE_VERSION) \
	-f update.Dockerfile \
	--platform linux/amd64,linux/arm64 \
	--push .
