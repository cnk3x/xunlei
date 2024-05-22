VERSION := $(shell ls spk | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g')
GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

GITTAG := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags '-s -w -X main.version=$(GITTAG)'
BUILD_FLAGS :=-trimpath -v $(LDFLAGS)

showTag:
	@echo $(GITTAG)

build-amd64::
	rm -f bin/xlp-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/xlp-amd64 ./cmd/xlp

build-arm64::
	rm -f bin/xlp-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/xlp-arm64 ./cmd/xlp

build-amd64-embed::
	rm -f bin/xlp-amd64-embed
	cp -f embeds/nasxunlei-amd64.rpk embeds/nasxunlei.rpk 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -tags embed -v -o bin/xlp-amd64-embed ./cmd/xlp

build-arm64-embed::
	rm -f bin/xlp-arm64-embed
	cp -f embeds/nasxunlei-arm64.rpk embeds/nasxunlei.rpk 
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -tags embed -v -o bin/xlp-arm64-embed ./cmd/xlp

home:: build-amd64
	docker buildx build --push --platform linux/amd64 \
	-t $(shell cat home.repo.txt)/xunlei:latest \
	.
