VERSION:=$(shell ls spk | head -n1 | grep v | grep $(uname -m).spk | grep -Eo 'v[0-9]+.[0-9]+.[0-9]' | sed 's/v//g')

version:
	@echo $(VERSION)

build:
	docker buildx build -t cnk3x/xunlei:latest --load .

push:
	docker buildx build -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform linux/amd64,linux/arm64 --push .
