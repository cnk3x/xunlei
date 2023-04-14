[![docker hub pulls](https://img.shields.io/docker/pulls/cnk3x/xunlei.svg)](https://hub.docker.com/r/cnk3x/xunlei) [![docker hub version](https://img.shields.io/docker/v/cnk3x/xunlei)](https://hub.docker.com/r/cnk3x/xunlei) [![GitHub Repo stars](https://img.shields.io/github/stars/cnk3x/xunlei)](https://github.com/cnk3x/xunlei) 



# 迅雷远程下载服务(docker)(非官方)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。仅供测试，测试完请大家自觉删除。

下载保存目录 `/xunlei/downloads`， 对应迅雷应用内显示的下载路径是 `/downloads` 或者 `/迅雷下载`

[容器镜像: cnk3x/xunlei](https://hub.docker.com/r/cnk3x/xunlei)

[阿里云镜像（国内访问）: registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:latest](#)

[源码仓库: https://github.com/cnk3x/xunlei/tree/docker](https://github.com/cnk3x/xunlei/tree/docker)

- 环境变量 `XL_WEB_PORT`: 网页访问端口，默认 `2345`。
- 环境变量 `XL_WEB_ADDRESS` 绑定端口，默认 `:port`
- 环境变量 `XL_DEBUG`: 1 为调试模式，输出详细的日志信息，0: 关闭，不显示迅雷套件输出的日志，默认0.
- 环境变量 `UID`, `GID`, 设定运行迅雷下载的用户，使用此参数注意下载目录必须是该账户有可写权限。 https://github.com/cnk3x/xunlei/issues/85
- 环境变量 `XL_BA_USER` 和 `XL_BA_PASSWORD`: 给迅雷面板添加基本验证（明码）。 https://github.com/cnk3x/xunlei/issues/57
- `host` 网络下载速度比 `bridge` 快, 如果没有条件使用host网络，映射`XL_WEB_PORT`设定的端口`tcp`即可。
- 下载保存目录 `/xunlei/downloads`, 数据目录：`/xunlei/data`, 请持久化。
- `hostname`: 迅雷会以主机名来命名远程设备，你在迅雷App上看到的就是这个。
- ~~安装好绑定完后可以在线升级到迅雷官方最新版本~~ 这点不确定了，得自己尝试。

## docker shell

```bash
# 以下以 /mnt/sdb1/downloads 为实际的下载保存目录 /mnt/sdb1/xunlei 为实际的数据保存目录 为例
# 根据实际情况更改
# 如果已经安装过的(/mnt/sdb1/xunlei 目录已存在)，再次安装会复用，而且下载目录不可更改，如果要更改下载目录，请把这个目录删掉重新绑定。

# 国内访问将 cnk3x/xunlei:latest 替换为 registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:latest

# host网络，默认端口 2345
docker run -d --name=xunlei --hostname=mynas --net=host -v /mnt/sdb1/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest

# host网络，更改端口为 4321
docker run -d --name=xunlei --hostname=mynas --net=host -e XL_WEB_PORT=4321 -v /mnt/sdb1/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest

# bridge 网络，默认端口 2345
docker run -d --name=xunlei --hostname=mynas --net=bridge -p 2345:2345 -v /mnt/sdb1/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest

# bridge 网络，更改端口为 4321
docker run -d --name=xunlei --hostname=mynas --net=bridge -p 4321:2345 -v /mnt/sdb1/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest
```

## docker compose

```yaml
# host默认端口 2345
# compose.yml
services:
  xunlei:
    image: cnk3x/xunlei:latest
    privileged: true
    container_name: xunlei
    hostname: mynas
    network_mode: host
    volumes:
      - /mnt/sdb1/xunlei:/xunlei/data
      - /mnt/sdb1/downloads:/xunlei/downloads
    restart: unless-stopped
```

```yaml
# host更改端口 4321
# compose.yml
services:
  xunlei:
    image: cnk3x/xunlei:latest
    privileged: true
    container_name: xunlei
    hostname: mynas
    network_mode: host
    environment:
      - XL_WEB_PORT=4321
    volumes:
      - /mnt/sdb1/xunlei:/xunlei/data
      - /mnt/sdb1/downloads:/xunlei/downloads
    restart: unless-stopped
```

```yaml
# bridge默认端口 2345
# compose.yml
services:
  xunlei:
    image: cnk3x/xunlei:syno
    privileged: true
    container_name: xunlei
    hostname: mynas
    network_mode: bridge
    ports:
      - 2345:2345
    volumes:
      - /mnt/sdb1/xunlei:/xunlei/data
      - /mnt/sdb1/downloads:/xunlei/downloads
    restart: unless-stopped
```

```yaml
# bridge更改端口 4321
# compose.yml
services:
  xunlei:
    image: cnk3x/xunlei:syno
    privileged: true
    container_name: xunlei
    hostname: mynas
    network_mode: bridge
    ports:
      - 4321:2345
    volumes:
      - /mnt/sdb1/xunlei:/xunlei/data
      - /mnt/sdb1/downloads:/xunlei/downloads
    restart: unless-stopped
```

## 镜像构建流程
1. 前往迅雷NAS版本官网下载最新版本的迅雷套件，下载地址：https://nas.xunlei.com/
2. 将1中下载的迅雷套件放置于`spk`目录下
3. 执行`docker build -t xxx/xunlei --build-arg TARGETARCH={amd64|arm64} .`构建镜像

## 已知问题

插件无法使用

## systemd 服务版本

<https://github.com/cnk3x/xunlei/tree/main>

## Used By

[kubespider](https://github.com/jwcesign/kubespider/blob/main/docs/zh/user_guide/thunder_install_config/README.md)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=cnk3x/xunlei&type=Date)](https://star-history.com/#cnk3x/xunlei&Date)

