VERSION:=2.8.0

build:
	docker buildx build --target=vip -t cnk3x/xunlei:latest --load .
	docker buildx build --target=syno -t cnk3x/xunlei:syno --load .

push:
	docker buildx build --target=vip -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform=linux/amd64,linux/arm64 --push .
	docker buildx build --target=syno -t cnk3x/xunlei:syno-$(VERSION) -t cnk3x/xunlei:syno --platform=linux/amd64,linux/arm64 --push .
