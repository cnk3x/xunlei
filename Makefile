lsVERSION := $(shell ls spk | grep -Eo "v[0-9]+.[0-9]+.[0-9]+" | sort -Vr | head -n1 | sed 's/v//g')
GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

GITTAG := $(shell git describe --tags --always)
LDFLAGS := -ldflags '-s -w -X main.version=$(GITTAG)'
BUILD_FLAGS := -v $(LDFLAGS)
GO_BUILD := CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAGS)

DOCKER_BUILD := docker buildx build --push --platform linux/amd64,linux/arm64

showTag:
	@echo $(GITTAG)

amd64::
	rm -f bin/xlp-amd64*
	GOARCH=amd64 $(GO_BUILD) -tags embed -v -o bin/xlp-amd64 ./cmd/xlp

wsl::
	rm -f /usr/local/bin/xlp
	GOARCH=amd64 $(GO_BUILD) -tags embed -v -o /usr/local/bin/xlp ./cmd/xlp

arm64::
	rm -f bin/xlp-arm64*
	GOARCH=arm64 $(GO_BUILD) -tags embed -v -o bin/xlp-arm64 ./cmd/xlp

build:: amd64 arm64

home:: amd64
	docker buildx build --push -t $(shell cat home.repo.txt)/xunlei:latest -f docker/Dockerfile .

latestPush:: build
	$(DOCKER_BUILD) -t $(HUB)/xunlei:latest -t $(GHR)/xunlei:latest -t $(ALIR)/xunlei:latest -f docker/Dockerfile .

versionedPush:: build
	$(DOCKER_BUILD) -t $(HUB)/xunlei:$(VERSION) -t $(GHR)/xunlei:$(VERSION) -t $(ALIR)/xunlei:$(VERSION) -f docker/Dockerfile .

binary:: build
	rm -f bin/xlp bin/xlp-amd64.tar.gz
	cp bin/xlp-amd64 bin/xlp
	tar -zcf bin/xlp-amd64.tar.gz -C bin xlp
	rm -f bin/xlp bin/xlp-arm64.tar.gz
	cp bin/xlp-arm64 bin/xlp
	tar -zcf bin/xlp-arm64.tar.gz -C bin xlp
	rm -f bin/xlp

hello::
	echo $(HELLO) $(VERSION)

scp:: amd64
	scp ./bin/xlp-amd64 root@192.168.99.12:/usr/local/bin/xlp
