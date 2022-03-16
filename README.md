# 迅雷远程下载服务

## 说明

从迅雷群晖套件中提取出来用于其他Linux设备的迅雷远程下载服务程序。

本程序 **不要在群晖的机器上运行！** **不要在群晖的机器上运行！** **不要在群晖的机器上运行！** 群晖的机器使用迅雷官方提供的套件即可

## 一键安装

```sh
# 安装
sh -c "$(curl -fSsL https://mirror.ghproxy.com/https://raw.githubusercontent.com/cnk3x/xunlei/main/install.sh)" - install --port=2345 --download-dir=/download
# 上面命令后面的参数 --port=后面接端口号, --download-dir=接下载文件夹，按自己的需求改
# 下载文件夹装好后没得改了，要改的话，卸载重装，或者用软链接
# 有时候安装失败，可以先运行卸载一次，再安装
# 启动后，浏览器访问你的设备地址+端口号绑定迅雷就可以了。
# 当前版本支持迅雷官方公测前的在线更新(不需要重新状态)

# 卸载
sh -c "$(curl -fSsL https://mirror.ghproxy.com/https://raw.githubusercontent.com/cnk3x/xunlei/main/uninstall.sh)"

# 卸载旧版本v2.1.x
sh -c "$(curl -fSsL https://mirror.ghproxy.com/https://raw.githubusercontent.com/cnk3x/xunlei/main/uninstall_old.sh)"

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
