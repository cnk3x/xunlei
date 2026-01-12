创建 github action，含以下步骤和要求：

1. git 拉取最新的 tag 的代码库, 仓库: https://github.com/cnk3x/xunlei
2. 从代码读取版本号, 参考命令: `cat xlp.go | grep "const Version =" | head -n1 | grep -Eo '"[^"]+"' | sed 's/"//g'`
3. 编译 golang 程序, main 入口在代码库根目录的子目录/cmd/xlp, 静态编译, 只需要 linux/amd64 和 linux/arm64, 输出到产物 xlp-arm64 和 xlp-amd64
4. 将产物 xlp-arm64 和 xlp-amd64 分别改名成 xlp 并打包成 xlp-arm64.tar.gz 和 xlp-amd64.tar.gz，打包后改回原名称以备 Dockerfile 中使用
5. 推送一个 release，包含产物 xlp-arm64.tar.gz 和 xlp-amd64.tar.gz
6. 代码库根目录有一个 Dockerfile, 用此 Dockerfile 构建多架构镜像并推送
    1. 架构只需要 linux/amd64 和 linux/arm64
    2. 镜像目标
        1. cnk3x/xunlei:版本号
        2. ghcr.io/cnk3x/xunlei:版本号
        3. registry.cn-shenzhen.aliyuncs.com/cnk3x/cnk3x/xunlei:版本号
    3. 镜像仓库 push 用的账号和密码需要安全的方式提供
7. 支持手动触发和 tag 触发
8. secrets 命名:
    1. ACR_USER: 阿里云账号
    2. ACR_TOKEN: 阿里云 token
    3. DCR_USER: docker.io 账号
    4. DCR_TOKEN: docker.io token
