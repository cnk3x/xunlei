# 迅雷远程下载服务(非官方)

[![Docker Pulls](https://img.shields.io/docker/pulls/cnk3x/xunlei.svg)](https://hub.docker.com/r/cnk3x/xunlei)
[![Docker Version](https://img.shields.io/docker/v/cnk3x/xunlei)](https://hub.docker.com/r/cnk3x/xunlei)
[![GitHub Stars](https://img.shields.io/github/stars/cnk3x/xunlei)](https://star-history.com/#cnk3x/xunlei&Date)

从迅雷群晖套件中提取出来用于其他设备的迅雷远程下载服务程序。仅供研究学习测试。 \
本程序仅提供Linux模拟和容器化运行环境，未对原版迅雷程序进行任何修改。

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
XL_DASHBOARD_USERNAME  #网页访问的用户名
XL_DASHBOARD_PASSWORD  #网页访问的密码
XL_DIR_DOWNLOAD        #下载保存默认文件夹，默认 /xunlei/downloads
XL_DIR_DATA            #程序数据保存文件夹，默认 /xunlei/data
XL_LOG                 #日志文件输出目标，默认为 null, 可选 file, console
XL_LOGGER_MAXSIZE      #日志文件最大大小
XL_LOGGER_COMPRESS     #是否压缩日志文件
```

#### 在容器中运行

```bash
# docker run -d \
#   -v <数据目录>:/data \
#   -v <默认下载保存目录>:/downloads \
#   -p <访问端口>:2345 \
#   cnk3x/xunlei

# example
docker run -d -v  /mnt/sdb1/xunlei:/xunlei/data -v /mnt/sdb1/downloads:/xunlei/downloads -p 2345:2345 cnk3x/xunlei
```

也可以直接运行

```bash
Usage of xlp:
  -dashboard-host string
        网页控制台访问绑定主机或IP, 不明白留空即可
  -dashboard-password string
        网页控制台访问密码
  -dashboard-port int
        网页控制台访问端口，默认 2345
  -dashboard-user string
        网页控制台访问用户名
  -dir-data string
        迅雷程序数据保存文件夹，默认 /xunlei/data
  -dir-download string
        默认下载保存文件夹，默认 /xunlei/downloads
  -log string
        日志输出位置, 可选 null, console, file
  -log-compress
        日志文件是否压缩
  -log-maxsize string
        日志文件最大大小
```

## 更新

重构了外壳程序，Docker镜像的基础包升级到 v3.12.0

## Used By

[kubespider](https://github.com/opennaslab/kubespider/blob/main/docs/zh/user_guide/thunder_install_config/README.md)
