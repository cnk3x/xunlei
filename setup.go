package main

import (
	"embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed target
var targetFiles embed.FS

//go:embed host
var hostFiles embed.FS

//go:embed xunlei-from-syno.service
var serviceFile []byte

const serviceFn = "/etc/systemd/system/xunlei-from-syno.service"

func install([]string) {
	copyBin := func() error {
		if me, _ := os.Executable(); me != "" {
			if err := FileCopy(me, filepath.Join(SYNOPKG_PKGBASE, "xunlei-from-syno")); err != nil {
				return err
			}
		}
		return nil
	}

	log.Printf("[安装]")
	printError("[安装]",
		DumpFs(targetFiles, "target", filepath.Join(SYNOPKG_PKGBASE, "target")),
		DumpFs(hostFiles, "host", filepath.Join(SYNOPKG_PKGBASE, "host")),
		os.Chmod(filepath.Join(SYNOPKG_PKGBASE, "target/bin/bin/xunlei-pan-cli-launcher.amd64"), 0755),
		os.Chmod(filepath.Join(SYNOPKG_PKGBASE, "target/bin/bin/xunlei-pan-cli.2.1.0.amd64"), 0755),
		os.Chmod(filepath.Join(SYNOPKG_PKGBASE, "target/ui/index.cgi"), 0755),
		os.Chmod(filepath.Join(SYNOPKG_PKGBASE, "host/usr/syno/synoman/webman/modules/authenticate.cgi"), 0755),
		copyBin(),
		os.WriteFile(serviceFn, serviceFile, 0644),
	)

	printError("[安装]",
		Shell("[安装]", "systemctl", "daemon-reload"),
		Shell("[安装]", "systemctl", "enable", "xunlei-from-syno"),
		Shell("[安装]", "systemctl", "start", "xunlei-from-syno"),
		Shell("[安装]", "systemctl", "status", "xunlei-from-syno"),
	)

	log.Printf("[安装]完成")
	log.Println()
	log.Printf("浏览器输入 http://你的IP:2345 使用迅雷")
}

func clean([]string) {
	log.Printf("[清理]")
	printError("[清理]",
		delFile(synoAuthenticatePath),
		delFile(synoInfoPath),
		Shell("[清理]", "systemctl", "stop", "xunlei-from-syno"),
		Shell("[清理]", "systemctl", "disable", "xunlei-from-syno"),
	)

	log.Printf("[清理] 请手动删除 %s", serviceFn)
	log.Printf("[清理] 请手动删除 %s", SYNOPKG_PKGBASE)
	log.Printf("[清理] 完成")
}

func printError(prefix string, errs ...error) {
	for _, err := range errs {
		if err != nil {
			log.Printf("%s %v", prefix, err)
		}
	}
}

func Shell(prefix string, name string, args ...string) error {
	bytes, err := exec.Command(name, args...).CombinedOutput()
	if len(bytes) > 0 {
		log.Printf("%s %s", prefix, bytes)
	}
	return err
}
