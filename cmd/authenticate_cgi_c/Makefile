TARGET := ../../embed/authenticate_cgi/authenticate_cgi_linux_

build::
	gcc main.c -o $(TARGET)amd64 -O2 -s
	aarch64-linux-gnu-gcc main.c -o $(TARGET)arm64 -O2 -s

build_static::
	gcc main.c -o $(TARGET)amd64-static -static -O2 -s
	aarch64-linux-gnu-gcc main.c -o $(TARGET)arm64-static -static -O2 -s
