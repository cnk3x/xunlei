# 迅雷远程下载服务

## 使用说明
> 从迅雷群晖套件中提取出来用于其他Linux设备的迅雷远程下载服务程序。

本程序不要在群晖的机器上运行，不要在群晖的机器上运行，不要在群晖的机器上运行！群晖的机器使用迅雷官方提供的套件即可

## 安装

```sh
# 安装
sh -c "$(curl -fSsL https://gh.k3x.cn/raw/cnk3x/xunlei/main/install.sh)" - install --port=2345 --download-dir=/download

# 卸载
sh -c "$(curl -fSsL https://gh.k3x.cn/raw/cnk3x/xunlei/main/install.sh)" - uninstall

# 有时候安装失败，可以先运行卸载一次，再安装
```
