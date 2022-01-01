package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed target
var targetFiles embed.FS

const (
	appname              = "xunlei"
	unitFile             = "/etc/systemd/system/" + appname + ".service"
	synoAuthenticatePath = "/usr/syno/synoman/webman/modules/authenticate.cgi"
	synoInfoPath         = "/etc/synoinfo.conf"
)

// 安装
type Install struct {
	Description string `flag:"description" alias:"d" usage:"服务描述" default:"迅雷远程下载服务"`
	Port        int    `flag:"port" alias:"p" env:"XUNLEI_PORT" usage:"监听端口" default:"2345"`
	Internal    bool   `flag:"internal" env:"XUNLEI_INTERNAL" usage:"仅本机访问，如果是通过反向代理来访问可选此项"`
	DownloadDIR string `flag:"download-dir" alias:"dir" env:"XUNLEI_DOWNLOAD_DIR" usage:"下载保存目录" default:"/downloads"`
}

func (c *Install) Usage() string {
	return "安装"
}

func (c *Install) Run(ctx context.Context, args []string) error {
	return contextRun(ctx,
		c.extract, // 解压文件
		c.config,  // 配置
		c.service, // 安装服务
	)
}

// 安装文件
func (c *Install) extract(ctx context.Context) error {
	// checkShellPID(getShellPID(), true)
	log := Standard("安装")

	var (
		targetDIR     = filepath.Join(SYNOPKG_PKGBASE, "target")
		hostDIR       = filepath.Join(targetDIR, "host")
		startEndpoint = filepath.Join(SYNOPKG_PKGBASE, appname)
	)

	// 释放文件
	{
		log.Infof("释放文件")
		if err := dumpFs(targetFiles, "target", targetDIR, log); err != nil {
			log.Fatalf("不成功: %v", err)
		}

		rb := make([]byte, 32)
		rand.Read(rb)
		rs := hex.EncodeToString(rb)[:7]
		f0 := filepath.Join(SYNOPKG_PKGDEST, "host", synoInfoPath)
		err := os.MkdirAll(filepath.Dir(f0), 0755)
		if err != nil {
			log.Fatalf("不成功: %v", err)
		}
		err = os.WriteFile(f0, []byte(`unique="synology_`+rs+`_720+"`), 0644)
		if err != nil {
			log.Fatalf("不成功: %v", err)
		}
		f1 := filepath.Join(SYNOPKG_PKGDEST, "host", synoAuthenticatePath)
		err = os.MkdirAll(filepath.Dir(f1), 0755)
		if err != nil {
			log.Fatalf("不成功: %v", err)
		}
		err = os.WriteFile(f1, []byte("#!/usr/bin/env sh\necho OK"), 0755)
		if err != nil {
			log.Fatalf("不成功: %v", err)
		}

		linkSyno := func(p string) (err error) {
			if _, err = os.Stat(p); err != nil {
				if os.IsNotExist(err) {
					if err = os.MkdirAll(filepath.Dir(p), 0755); err != nil {
						log.Fatalf("不成功: %v", err)
					}
					err = os.Symlink(filepath.Join(hostDIR, p), p)
				}
			}
			return
		}

		if err := linkSyno(synoAuthenticatePath); err != nil {
			log.Fatalf("不成功: %v", err)
		}

		if err := linkSyno(synoInfoPath); err != nil {
			log.Fatalf("不成功: %v", err)
		}

		if src, _ := os.Executable(); src != "" {
			log.Infof("  [Extract] %s", startEndpoint)
			if err := copyFile(src, startEndpoint); err != nil {
				log.Fatalf("  不成功: %v", err)
			}
		}

		log.Infof("释放完成")
		log.Infof("")
	}
	return nil
}

// 配置
func (c *Install) config(ctx context.Context) error {
	if c.Port == 0 {
		c.Port = 2345
	}
	if c.DownloadDIR == "" {
		c.DownloadDIR = "/downloads" ///mnt/nas/1
	}

	log := Standard("配置")
	log.Infof("网页端口: %d", c.Port)
	log.Infof("下载目录: %s", c.DownloadDIR)
	log.Infof("")

	cfg := []byte(fmt.Sprintf(`{"port":%d, "internal": %t, "dir":"%s"}`, c.Port, c.Internal, c.DownloadDIR))
	return os.WriteFile(filepath.Join(SYNOPKG_PKGBASE, "config.json"), cfg, 0666)
}

// 安装服务
func (c *Install) service(ctx context.Context) error {
	sUnit := fmt.Sprintf(`[Unit]
Description=%s
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=simple
ExecStart=/var/packages/pan-xunlei-com/xunlei run
LimitNOFILE=1024
LimitNPROC=512

[Install]
WantedBy=multi-user.target`, c.Description)

	if err := os.WriteFile(unitFile, []byte(sUnit), 0666); err != nil {
		return err
	}
	_ = serviceControlSilence("daemon-reload")(ctx)
	_ = serviceControlSilence("enable", appname)(ctx)
	return nil
}

// 卸载
type Uninstall struct {
	log Logger
}

func (c *Uninstall) Usage() string {
	return "卸载"
}

func (c *Uninstall) Run(ctx context.Context, args []string) error {
	c.log = Standard("卸载")
	return contextRun(ctx,
		serviceControlSilence("disable", appname),
		serviceControlSilence("stop", appname),
		c.removeUnit,
		serviceControlSilence("daemon-reload"),
		c.removeFiles,
	)
}

func (c *Uninstall) removeUnit(ctx context.Context) error {
	c.log.Infof("删除服务文件: %s", unitFile)
	return os.RemoveAll(unitFile)
}

func (c *Uninstall) removeFiles(ctx context.Context) error {
	c.log.Infof("删除迅雷文件: %s", SYNOPKG_PKGBASE)
	return os.RemoveAll(SYNOPKG_PKGBASE)
}
