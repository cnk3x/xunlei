# 迅雷群晖提取版

## 使用说明
> 从迅雷群晖套件中提取出来用于其他Linux设备的迅雷远程下载服务程序。

## 安装

```sh
# 进入临时目录
cd /tmp/
# 下载并解包
wget -O - https://github.com/shuxs/xunlei-from-symo/releases/download/v2.1.0/xunlei-from-syno.x86_64.tar.gz | tar zx
# 安装
./xunlei-from-syno install && rm -f ./xunlei-from-syno
# 安装完成之后就可以用 http://你设备的IP:2345 来访问了。
```

[也可以下载编译好的程序自行琢磨](https://github.com/shuxs/xunlei-from-symo/releases/download/v2.1.0/xunlei-from-syno.v2.1.0.x86_64.tar.gz) 

## 控制

```sh
# 启动
systemctl start xunlei-from-syno
# 停止
systemctl stop xunlei-from-syno
# 状态
systemctl status xunlei-from-syno
```

## 过程
> 尽可能的不污染系统, 安装过程她会做以下处理

1. 创建 `/var/packages/pan-xunlei-com` 目录，在里面释放所需的文件和配置。
2. 创建一个名为 `xunlei-from-syno` 的服务，服务文件路径 `/etc/systemd/system/xunlei-from-syno.service` 
3. 启动后会写一些文件。 
    - `/usr/syno/synoman/webman/modules/authenticate.cgi`
    - `/etc/synoinfo.conf`
4. 设置 `xunlei-from-syno` 服务开机自启
5. 启动 `xunlei-from-syno` 服务

> 默认配置，请确保不会冲突
1. 下载目录: `/downloads`, 如需改用其他目录，请自行编辑服务文件
2. 网页端口: `2345`, 如需改用其他目录，请自行编辑服务文件

```ini
# 服务文件
[Unit]
Description=迅雷群晖提取版
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=simple
ExecStart=/var/packages/pan-xunlei-com/xunlei-from-syno run --name U-NAS-迅雷 --port 2345 --download-dir=/downloads

TimeoutStopSec=5s
LimitNOFILE=1048576
LimitNPROC=512
PrivateTmp=true
ProtectSystem=full
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
```

## 交流

> QQQ: `238097122`