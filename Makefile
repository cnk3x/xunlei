GHR := ghcr.io/cnk3x
ALIR := registry.cn-shenzhen.aliyuncs.com/cnk3x
HUB := cnk3x

GITTAG := $(shell git describe --tags --always)
GO_BUILD := CGO_ENABLED=0 GOOS=linux go build -v -ldflags '-s -w -X main.version=$(GITTAG)'

DOCKER_BUILD := docker buildx build --push --platform linux/amd64,linux/arm64

VERSION := $(GITTAG)

showTag:
	@echo $(GITTAG)

amd64::
	rm -f bin/xlp-amd64*
	GOARCH=amd64 $(GO_BUILD) -v -o bin/xlp-amd64 ./cmd/xlp

arm64::
	rm -f bin/xlp-arm64*
	GOARCH=arm64 $(GO_BUILD) -v -o bin/xlp-arm64 ./cmd/xlp

build:: amd64 arm64

latest:: build
	$(DOCKER_BUILD) -t $(HUB)/xunlei:latest -t $(GHR)/xunlei:latest -t $(ALIR)/xunlei:latest -f docker/Dockerfile .

versioned:: build
	$(DOCKER_BUILD) -t $(HUB)/xunlei:$(VERSION) -t $(GHR)/xunlei:$(VERSION) -t $(ALIR)/xunlei:$(VERSION) -f docker/Dockerfile .

push:: build 
	$(DOCKER_BUILD) -t $(HUB)/xunlei:$(VERSION) -t $(HUB)/xunlei:latest -t $(GHR)/xunlei:$(VERSION) -t $(GHR)/xunlei:latest -t $(ALIR)/xunlei:$(VERSION) -t $(ALIR)/xunlei:latest -f docker/Dockerfile .

binary:: build
	rm -f bin/xlp bin/xlp-amd64.tar.gz
	cp bin/xlp-amd64 bin/xlp
	tar -zcf bin/xlp-amd64.tar.gz -C bin xlp
	rm -f bin/xlp bin/xlp-arm64.tar.gz
	cp bin/xlp-arm64 bin/xlp
	tar -zcf bin/xlp-arm64.tar.gz -C bin xlp
	rm -f bin/xlp

wsl::
	rm -f /usr/local/bin/xlp
	GOARCH=amd64 $(GO_BUILD) -v -o /usr/local/bin/xlp ./cmd/xlp

home:: amd64
	docker buildx build --push -t $(shell cat home.repo)/xunlei:$(VERSION) -f docker/Dockerfile .

hello::
	echo $(HELLO) $(VERSION)

scp:: amd64
	scp ./bin/xlp-amd64 root@192.168.99.12:/usr/local/bin/xlp
