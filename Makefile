# makefile for dev test, production used github action to build
NAME := xlp

http_proxy := http://host.docker.internal:7890
https_proxy := http://host.docker.internal:7890

GBuild := CGO_ENABLED=0 GOOS=linux go build -v -ldflags '-s -w'
DProxy := --build-arg http_proxy=$(http_proxy) --build-arg https_proxy=$(https_proxy)
DBuild := docker buildx build
DPush := $(DBuild) --push --platform linux/amd64,linux/arm64

VERSION := $(shell cat xlp.go | grep "const Version =" | head -n1 | grep -Eo '"[^"]+"' | sed 's/"//g')

showTag::
	@echo version is $(VERSION)

amd64::
	GOOS=linux GOARCH=amd64 $(GBuild) -v -o artifacts/xlp-amd64 ./cmd/xlp
	cp artifacts/xlp-amd64 artifacts/xlp
	tar -C artifacts -czvf artifacts/xlp-$(VERSION)-linux-amd64.tar.gz xlp
	rm artifacts/xlp

arm64::
	GOOS=linux GOARCH=arm64 $(GBuild) -v -o artifacts/xlp-arm64 ./cmd/xlp
	cp artifacts/xlp-arm64 artifacts/xlp
	tar -C artifacts -czvf artifacts/xlp-$(VERSION)-linux-arm64.tar.gz xlp
	rm artifacts/xlp

build:: amd64 arm64

busybox:: amd64
	$(DBuild) --load -t $(NAME):$(VERSION) .

ubuntu:: amd64
	$(DBuild) --load -t $(NAME):$(VERSION)-ubuntu -f ubuntu.Dockerfile .

debian:: amd64
	$(DBuild) --load -t $(NAME):$(VERSION)-debian -f debian.Dockerfile .

test:: amd64
	wsl -d debian -- sshpass -p wan scp ./artifacts/xlp-amd64 cnk3x@192.168.99.9:~/apps/xunlei/xlp
