# 迅雷远程下载服务(docker)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。

x86_64 版本已在万由的U-NAS系统的Docker测试通过，arm64 没有机器，暂时未测。

[代码](https://github.com/cnk3x/xunlei/tree/docker)

[hub](https://hub.docker.com/r/cnk3x/xunlei)

## 安装

### docker shell

```bash
# 群晖迅雷模式运行
docker run -d --name=xunlei --hostname=my-nas-1 --net=host \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  -e=XL_WEB_PORT=2345 \
  --restart=always --privilage cnk3x/xunlei:latest syno

# Docker迅雷模式运行，需要白金以上会员，不需要 privilage 权限
# 如果是白金会员，推荐此方式
docker run -d --name=xunlei --hostname=my-nas-1 --net=host \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  -e=XL_WEB_PORT=2345 \
  --restart=always cnk3x/xunlei:latest

# bridge 网络请开放网页端口
docker run -d --name=xunlei --hostname=my-nas-1 --net=bridge \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  -e=XL_WEB_PORT=2345 \
  -p=2345:2345 \
  --restart=always cnk3x/xunlei:latest
```

### docker compose

```yaml
# compose.yml
services:
  xunlei:
    image: cnk3x/xunlei:latest
    # 取消注释下面两行以群晖迅雷方式运行, privilage 是必须的，否则需要白金会员
    # command: syno
    # privilage: true
    container_name: xunlei
    hostname: my-nas-1
    network_mode: host
    #network_mode: bridge
    #ports:
    #  - "2345:2345"
    #environment:
    #  XL_WEB_PORT=2345
    volumes:
      - <数据目录>:/data
      - <下载目录>:/downloads
    restart: always
```

### 说明

- `/downloads` 挂载点为下载目录。
- `/data` 挂载点为数据目录。
- `hostname` 迅雷会以主机名来命名远程设备，你在App上看到的就是这个。
- 设置网络模式为`host`下载速度和`bridge`网络模式不是一个量级的。
- 网页端口号通过环境变量 `XL_WEB_PORT` 来设置, 默认 2345
- 删除`<数据目录>`重启容器需要重新登录迅雷账号绑定
