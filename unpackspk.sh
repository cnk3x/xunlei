targetDIR=tmp/var/packages/pan-xunlei-com/target
mkdir -p ${targetDIR}
if [ "$(uname -m)" = "aarch64" ]; then arch=armv8; else arch=$(uname -m); fi
tar --wildcards -Oxf $(find spk -type f -name \*-${arch}.spk | head -n1) package.tgz | tar --wildcards -xJC ${targetDIR} 'bin/bin/*' 'ui/index.cgi'
mv ${targetDIR}/bin/bin/* ${targetDIR}/bin
rm -rf ${targetDIR}/bin/bin
