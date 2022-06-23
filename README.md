# 迅雷远程下载服务(docker)(非官方)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。

维护多个版本供选择比较麻烦，从2.9.1开始，只提供模拟群晖的版本。镜像名称为 [cnk3x/xunlei:latest](https://hub.docker.com/r/cnk3x/xunlei)

[源码仓库: https://github.com/cnk3x/xunlei/tree/docker](https://github.com/cnk3x/xunlei/tree/docker)

[容器镜像: cnk3x/xunlei](https://hub.docker.com/r/cnk3x/xunlei)

- 环境变量 `XL_WEB_PORT`: 网页访问端口，默认 `2345`。
- 环境变量 `XL_HOME`: 数据目录（保存迅雷账号，设置等信息），默认 `/data`。
- 环境变量 `XL_DOWNLOAD_PATH`: 下载保存根目录，默认 `/downloads`。
- 环境变量 `XL_DEBUG`: 1 为调试模式，输出详细的日志信息，0: 关闭，不显示迅雷套件输出的日志，默认0.
- `hostname`: 迅雷会以主机名来命名远程设备，你在迅雷App上看到的就是这个。
- `host` 网络下载速度比 `bridge` 快, 如果没有条件使用host网络，映射`XL_WEB_PORT`设定的端口`tcp`即可。
- 安装好绑定完后可以在线升级到迅雷官方最新版本

## docker shell

```bash
# 以下 
# /mnt/sdb1/downloads 为实际的下载保存目录
# /mnt/sdb1/docker/apps/xunlei/data 为实际的数据保存目录
# 根据实际情况更改
# 如果已经安装过的(/mnt/sdb1/docker/apps/xunlei/data 目录已存在)，再次安装会复用，而且下载目录不可更改，如果要更改下载目录，请把这个目录删掉重新绑定。

# host网络，默认端口 2345
docker run -d --name=xunlei --hostname=mynas --net=host -v /mnt/sdb1/docker/apps/xunlei/data:/data -v /mnt/sdb1/downloads:/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest

# host网络，更改端口为 4321
docker run -d --name=xunlei --hostname=mynas --net=host -e XL_WEB_PORT=4321 -v /mnt/sdb1/docker/apps/xunlei/data:/data -v /mnt/sdb1/downloads:/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest

# bridge 网络，默认端口 2345
docker run -d --name=xunlei --hostname=mynas --net=bridge -p 2345:2345 -v /mnt/sdb1/docker/apps/xunlei/data:/data -v /mnt/sdb1/downloads:/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest

# bridge 网络，更改端口为 4321
docker run -d --name=xunlei --hostname=mynas --net=bridge -p 4321:2345 -v /mnt/sdb1/docker/apps/xunlei/data:/data -v /mnt/sdb1/downloads:/downloads --restart=unless-stopped --privileged cnk3x/xunlei:latest
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
      - /mnt/sdb1/docker/apps/xunlei/data:/data
      - /mnt/sdb1/downloads:/downloads
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
      - /mnt/sdb1/docker/apps/xunlei/data:/data
      - /mnt/sdb1/downloads:/downloads
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
      - /mnt/d/docker/apps/xunlei/data:/data
      - /mnt/d/downloads:/downloads
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
      - /mnt/d/docker/apps/xunlei/data:/data
      - /mnt/d/downloads:/downloads
    restart: unless-stopped
```
