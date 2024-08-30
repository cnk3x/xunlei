VERSION=$(git describe --tags --always)

docker buildx build \
  -t cnk3x/xunlei:busybox \
  -t cnk3x/xunlei:${VERSION}-busybox \
  -t ghcr.io/cnk3x/xunlei:busybox \
  -t ghcr.io/cnk3x/xunlei:${VERSION}-busybox \
  -t registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:busybox \
  -t registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:${VERSION}-busybox \
  -f ./docker/busybox/Dockerfile \
  --platform "linux/amd64,linux/arm64" \
  --push .
