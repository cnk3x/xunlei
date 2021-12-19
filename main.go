package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	//go:embed host
	hostFiles embed.FS
	//go:embed target
	targetFiles embed.FS
)

const serviceName = "xunlei"

var (
	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	sMan        = &ServiceMan{
		Name:        serviceName,
		Description: "迅雷远程下载服务",
		Command:     filepath.Join(SYNOPKG_PKGBASE, serviceName) + " run -p 2345 -d /mnt/nas/1/downloads",
	}
)

func main() {
	// defer cancel()
	go func() {
		select {
		case <-ctx.Done():
			Standard("调试", os.Stderr).Debugf("退出: %v", ctx.Err())
		}
	}()
	flag.ErrHelp = errors.New("")
	app := cli.NewApp()
	app.Usage = "迅雷远程下载服务的安装和控制程序"
	app.Commands = []*cli.Command{
		commandInstall(),
		commandRun(),
		commandService(),
		commandUpgrade(),
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatalf("%v", err)
	}
}

func commandInstall() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "初始化安装",
		Action: func(c *cli.Context) error {
			checkShellPID(getShellPID(), true)

			log := Standard("安装")
			log.Infof("")

			var (
				targetDIR     = filepath.Join(SYNOPKG_PKGBASE, "target")
				hostDIR       = filepath.Join(targetDIR, "host")
				startEndpoint = filepath.Join(SYNOPKG_PKGBASE, sMan.Name)
			)

			// 释放文件
			{
				log.Infof("释放文件")
				if err := dumpFs(targetFiles, "target", targetDIR, log); err != nil {
					log.Fatalf("不成功: %v", err)
				}
				if err := dumpFs(hostFiles, "host", hostDIR, log); err != nil {
					log.Fatalf("不成功: %v", err)
				}
				if err := os.Chmod(filepath.Join(hostDIR, synoAuthenticatePath), 0755); err != nil {
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

			var (
				port         string
				downloadDIR  string
				runAsService string
			)

			if port == "" {
				port = "2345"
			}

			if downloadDIR == "" {
				downloadDIR = "/mnt/nas/1/downloads"
			}

			log.Infof("")
			log.Infof("网页端口: %d", port)
			log.Infof("下载目录: %s", downloadDIR)
			log.Infof("安装服务: %s", runAsService)
			log.Infof("")

			sMan.Command = fmt.Sprintf("%s run --port=%s --download-dir=%q", startEndpoint, port, downloadDIR)
			if err := sMan.Install(); err != nil {
				log.Fatalf("安装服务出错: %v", err)
			}
			if err := sMan.Start(); err != nil {
				log.Fatalf("启动服务出错: %v", err)
			}

			time.Sleep(time.Second * 3)
			if pid := getShellPID(); pid > 0 {
				if err := checkShellPID(getShellPID(), false); err == ErrPIDRunning {
					log.Infof("")
					log.Infof("已启动，PID: %d, 通过以下网址去绑定迅雷账号", pid)
					printIP("  http://%s:"+port, log)
					log.Infof("")
					return nil
				}
			}
			log.Infof("未能获取服务启动状态，请使用 systemctl status %s 命令查看服务状态", sMan.Name)
			return nil
		},
	}
}

func commandRun() *cli.Command {
	var options Options
	return &cli.Command{
		Name:  "run",
		Usage: "运行",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "download-dir",
				Aliases:     []string{"d"},
				Usage:       "下载目录",
				EnvVars:     []string{"XUNLEI_DOWNLOAD_DIR"},
				Value:       "/downloads",
				DefaultText: "/downloads",
				Destination: &options.DownloadPATH,
			},
			&cli.IntFlag{
				Name:        "web-port",
				Aliases:     []string{"port"},
				Usage:       "网页访问端口",
				EnvVars:     []string{"XUNLEI_WEB_PORT"},
				Value:       2345,
				DefaultText: "2345",
				Destination: &options.Port,
			},
			&cli.BoolFlag{
				Name:        "internal",
				Usage:       "是否仅允许本地访问",
				EnvVars:     []string{"XUNLEI_WEB_INTERNAL"},
				Destination: &options.Internal,
			},
		},
		Action: func(c *cli.Context) error {
			log := Standard("启动")
			log.Infof("面板端口: %d", options.Port)
			log.Infof("下载目录: %s", options.DownloadPATH)

			pidFn := filepath.Join(SYNOPKG_VAR, sMan.Name+"-shell.pid")
			os.WriteFile(pidFn, []byte(strconv.Itoa(os.Getpid())), 0600)
			defer os.Remove(pidFn)

			defer NewApp(options).Start().Stop()
			<-c.Context.Done()
			return nil
		},
	}
}

func commandUpgrade() *cli.Command {
	return &cli.Command{
		Name:  "upgrade",
		Usage: "更新自己",
		Action: func(c *cli.Context) error {
			log := Standard("更新")
			log.Infof("复制自己到Package目录")
			if src, _ := os.Executable(); src != "" {
				target := filepath.Join(SYNOPKG_PKGBASE, sMan.Name)
				log.Debugf("[Extract] %s", target)
				if err := copyFile(src, target); err != nil {
					return err
				}
			}
			log.Infof("完成")
			return nil
		},
	}
}

func commandService() *cli.Command {
	serviceControl := func(command, description string) *cli.Command {
		return &cli.Command{
			Name:  command,
			Usage: description,
			Action: func(c *cli.Context) error {
				switch command {
				case "start":
					return sMan.Start()
				case "stop":
					return sMan.Stop()
				case "install":
					log := Standard("安装服务")
					log.Infof("")
					if err := sMan.Install(); err != nil {
						return err
					}
					log.Infof("完成")
					log.Infof("可通过以下命令控制服务启停:")
					startEndpoint := filepath.Join(SYNOPKG_PKGBASE, sMan.Name)
					log.Infof("%s start   启动服务", startEndpoint)
					log.Infof("%s stop    停止服务", startEndpoint)
					log.Infof("%s disable 卸载服务", startEndpoint)
					return nil
				case "uninstall":
					log := Standard("卸载服务")
					log.Infof("")
					return sMan.Uninstall()
				default:
					return nil
				}
			},
		}
	}

	return &cli.Command{
		Name:            "service",
		Aliases:         []string{"srv"},
		Usage:           "服务安装和控制",
		HideHelp:        true,
		HideHelpCommand: true,
		Subcommands: []*cli.Command{
			serviceControl("start", "启动服务"),
			serviceControl("stop", "停止服务"),
			serviceControl("install", "安装服务"),
			serviceControl("uninstall", "卸载服务"),
		},
	}
}

func getShellPID() int {
	pidFn := filepath.Join(SYNOPKG_VAR, sMan.Name+"-shell.pid")
	data, _ := os.ReadFile(pidFn)
	pid, _ := strconv.Atoi(string(data))
	return pid
}
