# 迅雷远程下载服务(非官方)

## 说明

从迅雷群晖套件中提取出来用于其他Linux设备的迅雷远程下载服务程序。**已支持docker**

本程序 **不要在群晖的机器上运行！** **不要在群晖的机器上运行！** **不要在群晖的机器上运行！** 群晖的机器使用迅雷官方提供的套件即可

**提 issue 请一定要注明使用方式：docker 还是 本机服务运行， 是否root账号运行， arm64 还是 x86_64。**

## 一键安装

```sh
# 安装
sh -c "$(curl -fSsL https://raw.githubusercontent.com/cnk3x/xunlei/main/install.sh)" - install --port=2345 --download-dir=/download
# 上面命令后面的参数 --port=后面接端口号, --download-dir=接下载文件夹，按自己的需求改
# 下载文件夹装好后没得改了，要改的话，卸载重装，或者用软链接
# 有时候安装失败，可以先运行卸载一次，再安装
# 启动后，浏览器访问你的设备地址+端口号绑定迅雷就可以了。 比如： http://192.168.3.11:2345
# 当前版本支持迅雷官方公测前的在线更新(不需要重新安装)

# 卸载
sh -c "$(curl -fSsL https://raw.githubusercontent.com/cnk3x/xunlei/main/uninstall.sh)"

# 卸载旧版本v2.1.x
sh -c "$(curl -fSsL https://raw.githubusercontent.com/cnk3x/xunlei/main/uninstall_old.sh)"

# 服务控制
# 启动
systemctl start xunlei
# 停止
systemctl stop xunlei
# 状态
systemctl status xunlei
# 查看日志(ctrl+c退出日志查看)
journalctl -fu xunlei
```

## 更新

使用应用内更新的功能

## 自行编译

克隆源码
1. 下载[官方的对应架构的群晖版迅雷spk文件](https://docs.qq.com/doc/DQVJpbEVGZXV0anNa)
1. 用解压软件解压spk文件
1. 找到里面的 package.tgz, 再解压一次
1. 找到里面的文件: `xunlei-pan-cli-launcher`, `xunlei-pan-cli.版本号.amd64`, `index.cgi`
1. 找到里面的文件: 与 `xunlei-pan-cli.版本号.amd64` 同目录的version文件
1. 将 `index.cgi` 改名为 `xunlei-pan-cli-web`
1. 将这四个文件复制到源码target目录
1. 改你要改的，改完后编译。

## Docker

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
 
## 折腾了一个基于busybox的小镜像版本，可以尝试一下，x86_64没问题，arm64就不确定了，可能缺库。

用法与上面相同，把镜像改成下面即可

镜像:

cnk3x/xunlei:busybox

cnk3x/xunlei:busybox-syno
