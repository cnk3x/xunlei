#!/usr/bin/env sh

systemctl stop xunlei-from-syno
systemctl disable xunlei-from-syno
systemctl daemon-reload
rm -f "/etc/systemd/system/xunlei-from-syno.service"
rm -f "/usr/syno/synoman/webman/modules/authenticate.cgi"
rm -f "/etc/synoinfo.conf"
rm -fr "/var/packages/pan-xunlei-com"

echo 操作完成
