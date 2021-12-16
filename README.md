# 迅雷群晖提取版

## 使用说明
> 从迅雷群晖套件中提取出来用于其他Linux设备的迅雷远程下载服务程序。

## 安装

```sh
# 下载并解包
wget -O - https://github.com/shuxs/xunlei-from-syno/releases/download/v2.1.0/xunlei-from-syno.v2.1.0.x86_64.tar.gz | tar zx

# 国内可用
wget -O - https://github.91chi.fun//https://github.com//shuxs/xunlei-from-syno/releases/download/v2.1.0/xunlei-from-syno.v2.1.0.x86_64.tar.gz | tar zx
# 上面二者用其一即可

# 安装
./xunlei-from-syno install && rm -f ./xunlei-from-syno
# 安装完成之后就可以用 http://你设备的IP:2345 来访问了。

# 卸载
/var/packages/pan-xunlei-com/xunlei-from-syno clean
rm -rf /var/packages/pan-xunlei-com
rm -f /etc/systemd/system/xunlei-from-syno.service
# 除了你的下载文件夹之外，没有来过的痕迹了。
```

[也可以下载自行琢磨](https://github.com/shuxs/xunlei-from-syno/releases) 

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

## 配置
> 请确保不会冲突
1. 默认下载目录: `/downloads`
2. 默认网页端口: `2345`
3. 修改配置
    1. 停止服务 `systemctl stop xunlei-from-syno`
    2. 编辑服务 `vi /etc/systemd/system/xunlei-from-syno.service` 或者 `nano /etc/systemd/system/xunlei-from-syno.service`
    3. 修改 `ExecStart=/var/packages/pan-xunlei-com/xunlei-from-syno run --port=你要的网页端口 --download-dir=你要的下载目录`
    3. 重载服务 `systemctl daemon-reload`
    4. 启动服务 `systemctl start xunlei-from-syno`

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
## 关于Docker容器

> 代码中我写好 [Dockefile](https://github.com/shuxs/xunlei-from-syno/blob/main/Dockerfile) 和 [docker-compose.yaml](https://github.com/shuxs/xunlei-from-syno/blob/main/docker-compose.yaml), 能够跑起来，但登录账号后，输入邀请码的时候出现权限的提示，估计是需要docker平台的邀请码. 

```sh
docker compose build
docker compose up -d && docker compose logs -f
```

## 交流

> QQQ: `238097122`