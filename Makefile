DCR := docker.io/cnk3x/
ACR := registry.cn-shenzhen.aliyuncs.com/cnk3x/
GHCR := ghcr.io/cnk3x/
NAME := xlp

http_proxy := http://host.docker.internal:7890
https_proxy := http://host.docker.internal:7890

GoBuild := CGO_ENABLED=0 GOOS=linux go build -v -ldflags '-s -w'
DProxy := --build-arg http_proxy=$(http_proxy) --build-arg https_proxy=$(https_proxy)
DBuildBase := docker buildx build
DBuild := $(DBuildBase) --push --platform linux/amd64,linux/arm64

VERSION := $(shell cat xlp.go | grep "const Version =" | head -n1 | grep -Eo '"[^"]+"' | sed 's/"//g')

showTag::
	@echo version is $(VERSION)

amd64::
	GOARCH=amd64 $(GoBuild) -v -o artifacts/xlp-amd64 ./cmd/xlp

arm64::
	GOARCH=arm64 $(GoBuild) -v -o artifacts/xlp-arm64 ./cmd/xlp

build:: amd64 arm64

latest:: build
	$(DBuild) -t $(DCR)$(NAME):latest -t $(GHCR)$(NAME):latest -t $(ACR)$(NAME):latest .

tagged:: build
	$(DBuild) -t $(DCR)$(NAME):$(VERSION) -t $(GHCR)$(NAME):$(VERSION) -t $(ACR)$(NAME):$(VERSION) .

push:: build
	$(DBuild) -t $(DCR)$(NAME):$(VERSION) -t $(DCR)$(NAME):latest -t $(GHCR)$(NAME):$(VERSION) -t $(GHCR)$(NAME):latest -t $(ACR)$(NAME):$(VERSION) -t $(ACR)$(NAME):latest .

binary:: build
	@mv artifacts/xlp-amd64 artifacts/xlp
	@tar -zcf artifacts/xlp-amd64.tar.gz -C artifacts xlp
	@mv artifacts/xlp artifacts/xlp-amd64
	@mv artifacts/xlp-arm64 artifacts/xlp
	@tar -zcf artifacts/xlp-arm64.tar.gz -C artifacts xlp
	@mv artifacts/xlp artifacts/xlp-arm64
	@ls -lh artifacts

wsl::
	GOARCH=amd64 $(GoBuild) -v -o /usr/local/bin/xlp ./cmd/xlp

home:: amd64
	docker buildx build --push -t $(HOME_REPO)$(NAME):$(VERSION) .

load:: amd64
	docker buildx build --load -t $(NAME):$(VERSION) .

ubuntu::
	$(DBuildBase) --load -t $(NAME)-ubuntu:$(VERSION) -f ubuntu.Dockerfile .
