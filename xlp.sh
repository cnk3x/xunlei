uid=${XL_UID:"0"}
gid=${XL_GID:"0"}

if [ $gid -eq 0 ] && [ $uid -ne 0 ]; then gid=$uid; fi
if [ $uid -ne 0 ] && [ $gid -ne 0 ]; then addgroup -g $gid xunlei && adduser -D -h /data/.drive -u $uid -G xunlei xunlei; fi
