VERSION := $(shell ls spk | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g')
GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

version::
	@echo $(VERSION)

build::
	rm -f bin/xlp*
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-amd64 ./cmd/xlp
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-arm64 ./cmd/xlp

all:: build
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:$(VERSION)   -t $(GHR)/xunlei:latest \
	-t $(ALIR)/xunlei:$(VERSION)  -t $(ALIR)/xunlei:latest \
	-t $(HUB)/xunlei:$(VERSION)   -t $(HUB)/xunlei:latest \
	.

latestImage:: build
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:latest \
	-t $(ALIR)/xunlei:latest \
	-t $(HUB)/xunlei:latest \
	.

versionedImage:: build
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:$(VERSION) \
	-t $(ALIR)/xunlei:$(VERSION) \
	-t $(HUB)/xunlei:$(VERSION) \
	.

home:: build
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(shell cat home.repo.txt)/xunlei:$(VERSION) -t $(shell cat home.repo.txt)/xunlei:latest \
	.

testBuild::
	rm -f bin/xlp-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-amd64 ./cmd/xlp

local:: testBuild
	docker buildx build --load -t local/xunlei:latest .

localTest:: local
	echo $(shell pwd)
	docker run -it --rm --name xunlei \
		-v $(shell pwd)/testdata/downloads:/downloads \
		-v $(shell pwd)/testdata/data:/data \
		-e "XL_LOGGER_ENABLED=true" \
		-p 2346:2345 \
		local/xunlei:latest
