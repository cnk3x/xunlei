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

**常规**的容器，还是要在特权模式下运行。

如果 docker 的存储驱动如果是 btrfs 或者 overlayfs，可以支持的非特权运行。

#### 环境变量参数

```bash
XL_DASHBOARD_PORT      #网页访问的端口，默认 2345
XL_DASHBOARD_IP        #网页访问的端口，默认 0.0.0.0（代表所有IP）
XL_DASHBOARD_USERNAME  #网页访问的用户名
XL_DASHBOARD_PASSWORD  #网页访问的密码
XL_DIR_DOWNLOAD        #下载保存默认文件夹，默认 /xunlei/downloads，多个文件夹用冒号:分隔
XL_DIR_DATA            #程序数据保存文件夹，默认 /xunlei/data
XL_UID                 #运行迅雷的用户ID
XL_GID                 #运行迅雷的用户组ID
XL_PREVENT_UPDATE      #是否阻止更新，默认 true, 可选值 true/false, 1/0
XL_CHROOT              #隔离运行主目录, 指定该值且不为`/`则以隔离模式运行, 用于在容器内隔离环境，容器内默认为 /xunlei，隔离模式运行需要特权模式(--privileged)，可以将该值设置为`/`来以非特权模式运行。非特权模式运行有条件，可以尝试失败后使用特权模式重新运行。
XL_DEBUG               #调试模式, 可选值 true/false, 1/0
```

#### 在容器中运行

```bash
# docker run -d \
#   -v <数据目录>:/xunlei/data \
#   -v <默认下载保存目录>:/xunlei/downloads \
#   -p <访问端口>:2345 \
#   --privileged \
#   cnk3x/xunlei

# example
docker run --privileged -v /mnt/sdb1/configs/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads -p 2345:2345 cnk3x/xunlei

# 如果你的docker存储驱动是 overlayfs 或者 btrfs等, 可以不用特权运行
docker run -e XL_CHROOT=/ -v /mnt/sdb1/configs/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads -p 2345:2345 cnk3x/xunlei

```

也可以直接运行

```plain
$ bin/xlp-amd64 --help

Flags:
  -p, --dashboard_port      网页访问的端口 (env: XL_DASHBOARD_PORT) (default 2345)
  -i, --dashboard_ip        网页访问绑定IP，默认绑定所有IP (env: XL_DASHBOARD_IP)
  -u, --dashboard_username  网页访问的用户名 (env: XL_DASHBOARD_USERNAME)
  -k, --dashboard_password  网页访问的密码 (env: XL_DASHBOARD_PASSWORD)
      --dir_download        下载保存文件夹，可多次指定，需确保有权限访问 (env: XL_DIR_DOWNLOAD) (default [/xunlei/downloads])
      --dir_data            程序数据保存文件夹，其下'.drive'文件夹中，存储了登录的账号，下载进度等信息 (env: XL_DIR_DATA) (default "/xunlei/data")
      --uid                 运行迅雷的用户ID (env: XL_UID, UID)
      --gid                 运行迅雷的用户组ID (env: XL_GID, GID)
      --prevent_update      阻止更新 (env: XL_PREVENT_UPDATE) (default true)
  -r, --chroot              CHROOT主目录, 指定该值且不为/则以chroot模式运行, 用于在容器内隔离环境 (env: XL_CHROOT) (default "/")
      --debug               是否开启调试日志 (env: XL_DEBUG)
  -v, --version             显示版本信息
```

## Used By

[kubespider](https://github.com/opennaslab/kubespider/blob/main/docs/zh/user_guide/thunder_install_config/README.md)
