# 迅雷远程下载服务(docker)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。

x86_64 版本已在万由的U-NAS系统的Docker测试通过，arm64 没有机器，暂时未测。

[代码](https://github.com/cnk3x/xunlei/tree/docker)

[hub](https://hub.docker.com/r/cnk3x/xunlei)

## 安装

- `容器版迅雷模式`，需要白金以上会员，不需要 `privilage` 权限，如果是白金会员，推荐此方式。 
- `群晖版迅雷模式`，非会员每日三次机会，需要 `privilage` 权限。 
- `/downloads` 挂载点为下载目录。 `/data` 挂载点为数据目录。 
- `hostname` 迅雷会以主机名来命名远程设备，你在迅雷App上看到的就是这个。 
- `host` 网络下载速度比 `bridge` 快10倍。 
- `XL_WEB_PORT` 环境变量设置网页访问端口，默认 2345。
- `<数据目录>` 删除的话，重启容器需要重新登录迅雷账号绑定。

### docker shell

```bash
# 容器版迅雷模式
docker run -d --name=xunlei --hostname=my-nas-1 --net=host \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  # -e=XL_WEB_PORT=2345 \
  --restart=always cnk3x/xunlei:latest

# 群晖版迅雷模式
docker run -d --name=xunlei --hostname=my-nas-1 --net=host \
  -v=<数据目录>:/data -v=<下载目录>:/downloads \
  # -e=XL_WEB_PORT=2345 \
  --restart=always --privilage cnk3x/xunlei:latest syno
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
 