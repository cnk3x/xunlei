VERSION:=2.8.0

build:
	docker buildx build --target=vip -t cnk3x/xunlei:latest --load .
	docker buildx build --target=syno -t cnk3x/xunlei:syno --load .

push:
	docker buildx build --target=vip -t cnk3x/xunlei:$(VERSION) -t cnk3x/xunlei:latest --platform=linux/amd64,linux/arm64 --push .
	docker buildx build --target=syno -t cnk3x/xunlei:syno-$(VERSION) -t cnk3x/xunlei:syno --platform=linux/amd64,linux/arm64 --push .

build_busybox:
	docker buildx build --target=vip -t cnk3x/xunlei:busybox -f busybox.Dockerfile --load .
	docker buildx build --target=syno -t cnk3x/xunlei:syno-busybox -f busybox.Dockerfile --load .

push_busybox:
	docker buildx build --target=vip -t cnk3x/xunlei:busybox-$(VERSION) -t cnk3x/xunlei:busybox -f busybox.Dockerfile --platform=linux/amd64,linux/arm64 --push .
	docker buildx build --target=syno -t cnk3x/xunlei:syno-busybox-$(VERSION) -t cnk3x/xunlei:syno-busybox -f busybox.Dockerfile --platform=linux/amd64,linux/arm64 --push .

test_busybox:
	docker run --rm -it -p 2346:2345 --privileged cnk3x/xunlei:busybox sh