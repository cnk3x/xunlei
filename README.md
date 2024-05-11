# 迅雷远程下载服务(docker)(非官方)

[![Docker Pulls](https://img.shields.io/docker/pulls/cnk3x/xunlei.svg)](https://hub.docker.com/r/cnk3x/xunlei)
[![Docker Version](https://img.shields.io/docker/v/cnk3x/xunlei)](https://hub.docker.com/r/cnk3x/xunlei)
[![GitHub Stars](https://img.shields.io/github/stars/cnk3x/xunlei)](https://star-history.com/#cnk3x/xunlei&Date)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。仅供测试。

## 安装

当前支持容器中非特权运行。

### Docker

#### 镜像

```plain
cnk3x/xunlei:latest
registry.cn-shenzhen.aliyuncs.com/cnk3x/xunlei:latest
```

#### 环境变量参数

```bash
XL_DASHBOARD_PORT      #网页访问的端口
XL_DASHBOARD_HOST      #网页访问的地址
XL_DASHBOARD_USER      #网页访问的用户名
XL_DASHBOARD_PASSWORD  #网页访问的密码
XL_DIR_DOWNLOAD        #下载保存默认文件夹，默认 /downloads
XL_DIR_DATA            #程序数据保存文件夹，默认 /data
XL_LOG                 #日志文件输出目标，默认为 null, 可选 file, console
XL_LOGGER_MAXSIZE      #日志文件最大大小
XL_LOGGER_COMPRESS     #是否压缩日志文件
```

#### 安装命令

```bash
# docker run -d \
#   -v <数据目录>:/data \
#   -v <默认下载保存目录>:/downloads \
#   -p <访问端口>:2345 \
#   cnk3x/xunlei

# example
docker run -d -v /data:/data -v /downloads:/downloads -p 2345:2345 cnk3x/xunlei

```

## 更新

重构了外壳程序，Docker镜像的基础包升级到 v3.12.0

## Used By

[kubespider](https://github.com/opennaslab/kubespider/blob/main/docs/zh/user_guide/thunder_install_config/README.md)
