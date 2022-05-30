# 迅雷远程下载服务(docker)(非官方)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。

x86_64 版本已在万由的U-NAS系统的Docker测试通过，arm64 没有机器，暂时未测。

[源码: https://github.com/cnk3x/xunlei/tree/docker](https://github.com/cnk3x/xunlei/tree/docker)

[容器: cnk3x/xunlei](https://hub.docker.com/r/cnk3x/xunlei)

- 环境变量 `XL_WEB_PORT`: 网页访问端口，默认 `2345`。
- 环境变量 `XL_HOME`: 数据目录（保存迅雷账号，设置等信息），默认 `/data`。
- 环境变量 `XL_DOWNLOAD_PATH`: 下载保存根目录，默认 `/downloads`。
- 环境变量 `XL_DOWNLOAD_PATH_SUBS`: 如果有多个下载目录，使用次此参数来设置 (在群晖迅雷模式有效)。
- 环境变量 `XL_DEBUG`: 1 为调试模式，输出详细的日志信息，0: 关闭，不显示迅雷套件输出的日志，默认0.
- `容器版迅雷模式`: 需要白金以上会员，不需要 `privilage` 权限，如果是白金会员，推荐此方式。 
- `群晖版迅雷模式`: 非会员每日三次机会，需要 `privilage` 权限。 
- `hostname`: 迅雷会以主机名来命名远程设备，你在迅雷App上看到的就是这个。 
- `host` 网络下载速度比 `bridge` 快10倍, 如果没有条件使用host网络，映射`XL_WEB_PORT`设定的端口`tcp`即可 。 

### docker shell

```bash
# 容器版迅雷模式
docker run -d --name=xunlei --hostname=my-nas-1 --net=host \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  # -e=XL_WEB_PORT=2345 \
  --restart=always cnk3x/xunlei:latest
```

```bash
# 群晖版迅雷模式
docker run -d --name=xunlei --hostname=my-nas-1 --net=host \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  # -e=XL_WEB_PORT=2345 \
  --restart=always --privilage cnk3x/xunlei:syno
```

### docker compose

```yaml
# compose.yml
# 群晖迅雷模式
services:
  xunlei:
    image: cnk3x/xunlei:syno
    privilage: true
    container_name: xunlei
    hostname: my-nas-1
    network_mode: host
    #environment:
    #  XL_WEB_PORT=2345
    volumes:
      - <数据目录>:/data
      - <下载目录>:/downloads
    restart: always
```

```yaml
# compose.yml
# Docker迅雷模式
services:
  xunlei:
    image: cnk3x/xunlei:latest
    container_name: xunlei
    hostname: my-nas-1
    network_mode: host
    #environment:
    #  XL_WEB_PORT=2345
    volumes:
      - <数据目录>:/data
      - <下载目录>:/downloads
    restart: always
```
 