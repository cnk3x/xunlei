# 迅雷远程下载服务(非官方)

[![GitHub Stars][1]][2] [![Docker Pulls][3]][5] [![Docker Version][4]][5]

[1]: https://img.shields.io/github/stars/cnk3x/xunlei?style=flat
[2]: https://star-history.com/#cnk3x/xunlei&Date
[3]: https://img.shields.io/docker/pulls/cnk3x/xunlei.svg
[4]: https://img.shields.io/docker/v/cnk3x/xunlei
[5]: https://hub.docker.com/r/cnk3x/xunlei

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。仅供研究学习测试。 \
本程序仅提供 Linux 模拟和容器化运行环境，未对原版迅雷程序进行任何修改。

## 使用

### Docker

#### 镜像

```plain
cnk3x/xunlei:latest
registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:latest
ghcr.io/cnk3x/xunlei:latest
```

#### 参数

程序默认参数

```
OPTIONS:
      --dashboard_port uint16       网页访问的端口 [XL_DASHBOARD_PORT] (default 2345)
      --dashboard_ip ip             网页访问绑定IP，默认绑定所有IP [XL_DASHBOARD_IP]
      --dashboard_username string   网页访问的用户名 [XL_DASHBOARD_USERNAME]
      --dashboard_password string   网页访问的密码 [XL_DASHBOARD_PASSWORD]
  -d, --dir_download strings        下载保存文件夹，可多次指定，需确保有权限访问 [XL_DIR_DOWNLOAD] (default [./xunlei/downloads])
  -D, --dir_data string             程序数据保存文件夹，存储了登录的账号，下载进度等信息 [XL_DIR_DATA] (default "./xunlei/data")
  -u, --uid uint32                  运行迅雷的用户ID [XL_UID, UID]
  -g, --gid uint32                  运行迅雷的用户组ID [XL_GID, GID]
      --prevent_update              阻止更新 [XL_PREVENT_UPDATE]
  -r, --chroot string               CHROOT主目录 [XL_CHROOT] (default "./xunlei")
      --spk string                  SPK 下载链接 [XL_SPK_URL] (default "https://down.sandai.net/nas/nasxunlei-DSM7-x86_64.spk")
  -F, --force_download              强制下载
      --debug                       是否开启调试日志 [XL_DEBUG]
```

容器内参数默认值（在容器内运行时，覆盖程序默认参数）

```shell
#网页访问的端口
XL_DASHBOARD_PORT=2345
#网页访问绑定IP
XL_DASHBOARD_IP=
# 网页访问的用户名
XL_DASHBOARD_USERNAME=
# 网页访问的密码
XL_DASHBOARD_PASSWORD=
# 下载保存文件夹，多个用冒号`:`隔开
XL_DIR_DOWNLOAD=/xunlei/downloads
# 程序数据保存文件夹，存储了登录的账号，下载进度等信息
XL_DIR_DATA=/xunlei/data
# CHROOT主目录，这个不要改
XL_CHROOT=/xunlei
# 阻止更新
XL_PREVENT_UPDATE=true
# SPK下载链接
XL_SPK_URL=https://down.sandai.net/nas/nasxunlei-DSM7-(x86_64或者armv8).spk
# 运行迅雷的用户ID
XL_UID=
# 运行迅雷的用户GID
XL_GID=
# 是否开启调试日志
XL_DEBUG=
```

#### 在容器中运行

```bash
# 容器需要 SYS_ADMIN 权限运行，添加参数 --privileged 或者 --cap-add=SYS_ADMIN

# docker run -d \
#   -v <数据目录>:/xunlei/data \
#   -v <默认下载保存目录>:/xunlei/downloads \
#   -p <访问端口>:2345 \
#   --cap-add=SYS_ADMIN \
#   cnk3x/xunlei:latest

# example
docker run \
--name xunlei \
--hostname mynas \
-v /mnt/sdb1/configs/xunlei:/xunlei/data \
-v /mnt/sdb1/downloads:/xunlei/downloads \
-p 2345:2345 \
--cap-add=SYS_ADMIN \
cnk3x/xunlei
```

#### docker-compose

```yaml
services:
    xunlei:
        container_name: xunlei
        image: cnk3x/xunlei:latest
        restart: unless-stopped

        # 宿主机名，迅雷远程控制的名称与此相关，一般是 `群晖-${hostname}`
        hostname: my_storage
        # 必须, cap_add: [SYS_ADMIN] 和 privileged: true 二选一
        cap_add: [SYS_ADMIN]
        ports: [2345:2345] # 面板访问端口，如需更改，替换前面的2345即可
        environment:
            # 如果需要指定多个下载目录，手动指定XL_DIR_DOWNLOAD
            # 多个以冒号`:`隔开，都必须以 /xunlei 开头，迅雷面板选择保存路径显示会去掉/xunlei前缀
            # 指定后可以在 volumes 中绑定宿主机实际目录
            # 迅雷云盘的缓存会使用第一个目录会缓存
            # /xunlei/后面可以用中文
            # 不设置默认一个目录 /xunlei/downloads
            - XL_DIR_DOWNLOAD=/xunlei/downloads:/xunlei/movies:/xunlei/apps
        volumes:
            # 对应 XL_DIR_DOWNLOAD 指定的目录, 请替换冒号前面的路径为实际路径
            - 实际【下载】文件夹路径:/xunlei/downloads
            - 实际【电影】文件夹路径:/xunlei/movies
            - 实际【软件】文件夹路径:/xunlei/apps
            # 数据目录，必须，迅雷运行时，插件，升级，包括登录数据都在这
            - ./data:/xunlei/data
            # 可选，不配置每次启动都会从远程下载spk
            - ./cache:/xunlei/var/packages/pan-xunlei-com
```

## Used By

[kubespider](https://github.com/opennaslab/kubespider/blob/main/docs/zh/user_guide/thunder_install_config/README.md)
