TARGET := ../../embed/authenticate_cgi/authenticate_cgi_linux_
BUILD_S := CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags '-s -w' -o $(TARGET)

build::
	GOARCH=amd64 $(BUILD_S)amd64
	GOARCH=arm64 $(BUILD_S)arm64
