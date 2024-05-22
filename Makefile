VERSION := $(shell ls spk | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g')
GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

build::
	rm -f bin/xlp-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-amd64 ./cmd/xlp
	rm -f bin/xlp-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-arm64 ./cmd/xlp

buildEmbed::
	rm -f bin/xlp-amd64-embed
	cp -f embeds/nasxunlei-amd64.rpk embeds/nasxunlei.rpk 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -tags embed -v -o bin/xlp-amd64-embed ./cmd/xlp
	rm -f bin/xlp-arm64-embed
	cp -f embeds/nasxunlei-arm64.rpk embeds/nasxunlei.rpk 
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags '-s -w' -tags embed -v -o bin/xlp-arm64-embed ./cmd/xlp

versioned:: build
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:$(VERSION)  \
	-t $(ALIR)/xunlei:$(VERSION) \
	-t $(HUB)/xunlei:$(VERSION)  \
	.

latest:: build
	docker buildx build --push --platform linux/amd64,linux/arm64 \
	-t $(GHR)/xunlei:latest \
	-t $(ALIR)/xunlei:latest \
	-t $(HUB)/xunlei:latest \
	.

nasxunlei::
	rm -f bin/xlp-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-amd64 ./cmd/xlp
	docker buildx build --push -f nasxunlei.Dockerfile \
	-t $(GHR)/xunlei:latest \
	-t $(ALIR)/xunlei:latest \
	-t $(HUB)/xunlei:latest \
	.

home::
	@# linux/arm64
	rm -f bin/xlp-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-amd64 ./cmd/xlp
	docker buildx build --push --platform linux/amd64 \
	-t $(shell cat home.repo.txt)/xunlei:latest \
	.

testBuild::
	rm -f bin/xlp-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-s -w' -v -o bin/xlp-amd64 ./cmd/xlp

local:: testBuild
	docker buildx build --load -t local/xunlei:latest .

localTest:: local
	echo $(shell pwd)
	docker run -it --rm --name xunlei \
		-v $(shell pwd)/testdata/downloads:/xunlei/downloads \
		-v $(shell pwd)/testdata/data:/xunlei/data \
		-e "XL_LOGGER_ENABLED=true" \
		-p 2346:2345 \
		local/xunlei:latest
