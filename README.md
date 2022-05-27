# 迅雷远程下载服务(docker)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。

x86_64 版本已在万由的U-NAS系统的Docker测试通过，arm64 没有机器，暂时未测。

[代码](https://github.com/cnk3x/xunlei/tree/docker)

[hub](https://hub.docker.com/r/cnk3x/xunlei)

## 安装

### docker shell

```bash
docker run -d --name=xunlei \
  # 主机名，迅雷以此"群晖-主机名"来命名远程设备
  --hostname=my-nas-1 \
  # 设置为host下载会好一些
  # --net=host \
  # 开启调试模式会显示所有的迅雷日志，很多很多，正常使用时候不建议开启
  # -e XL_DEBUG=1 \
  -p=2345:2345 \
  -v=<数据目录>:/xunlei/data \
  -v=<下载目录>:/xunlei/downloads \
  --restart=always \
  cnk3x/xunlei:latest

docker run --rm --name=xunlei \
  --hostname=WenUNAS \
  -e XL_DEBUG=1 \
  -p=2345:2345 \
  -v=/mnt/nas/data/apps/xunlei/data:/data \
  -v=/mnt/nas/data/downloads:/xunlei/downloads \
  -v=/mnt/nas/data/media/:/xunlei/downloads/media \
  cnk3x/xunlei:latest
```

### docker compose

```yaml
# compose.yml
services:
  xunlei:
    image: cnk3x/xunlei:latest
    container_name: xunlei
    # 主机名，迅雷以此"群晖-主机名"来命名远程设备
    hostname: my-nas-1
    # 设置为host下载会快一些
    # network_mode: host
    network_mode: bridge
    ports:
      - "2345:2345"
    volumes:
      - ./data:/xunlei/data
      - <下载目录>:/xunlei/downloads
    restart: always
```

## 说明

默认网页管理端口 2345，如果要指定端口
- host网络: 通过指定命令 `docker run ... cnk3x/xunlei:latest xlp -port 5050` 改用 5050 端口
- bridge网络: 绑定时 `docker run ... -p 5050:2345 cnk3x/xunlei:latest` 改用 5050 端口

删除<数据目录>重启容器需要重新绑定
