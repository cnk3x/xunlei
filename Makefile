VERSION := $(shell ls spk | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g')
GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

GITTAG := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags '-s -w -X main.version=$(GITTAG)'
BUILD_FLAGS := -trimpath -v $(LDFLAGS)
GO_BUILD := CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS)

MULTI_BUILDX := docker buildx build --push --platform linux/amd64,linux/arm64

showTag:
	@echo $(GITTAG)

build-amd64::
	rm -f bin/xlp-amd64
	GOARCH=amd64 $(GO_BUILD) -o bin/xlp-amd64 ./cmd/xlp

build-arm64::
	rm -f bin/xlp-arm64
	GOARCH=arm64 $(GO_BUILD) -o bin/xlp-arm64 ./cmd/xlp

build-amd64-embed::
	rm -f bin/xlp-amd64-embed
	cp -f embeds/nasxunlei-amd64.rpk embeds/nasxunlei.rpk 
	GOARCH=amd64 $(GO_BUILD) -tags embed -v -o bin/xlp-amd64-embed ./cmd/xlp

build-arm64-embed::
	rm -f bin/xlp-arm64-embed
	cp -f embeds/nasxunlei-arm64.rpk embeds/nasxunlei.rpk 
	GOARCH=arm64 $(GO_BUILD) -tags embed -v -o bin/xlp-arm64-embed ./cmd/xlp

home:: build-amd64
	docker buildx build --push -t $(shell cat home.repo.txt)/xunlei:latest .

latestPush:: build-amd64 build-arm64
	$(MULTI_BUILDX) -t $(HUB)/xunlei:latest -t $(GHR)/xunlei:latest -t $(ALIR)/xunlei:latest .

versionedPush:: build-amd64 build-arm64
	$(MULTI_BUILDX) -t $(HUB)/xunlei:$(VERSION) -t $(GHR)/xunlei:$(VERSION) -t $(ALIR)/xunlei:$(VERSION) .