package spk

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cnk3x/xunlei/pkg/utils"
)

var DownloadUrl = utils.Iif(runtime.GOARCH == "amd64",
	"https://down.sandai.net/nas/nasxunlei-DSM7-x86_64.spk",
	"https://down.sandai.net/nas/nasxunlei-DSM7-armv8.spk",
)

// 检查并下载, 如果 force，忽略检查直接下载
func Download(ctx context.Context, spkUrl string, dir string, force bool) (err error) {
	if !force && allExists(ctx, dir) {
		slog.InfoContext(ctx, "check spk all spk file exists")
		return
	}

	switch {
	case utils.HasPrefix(spkUrl, "file://", true):
		err = download_file(ctx, spkUrl, dir)
	case utils.HasPrefix(spkUrl, "http://", true) || utils.HasPrefix(spkUrl, "https://", true):
		err = download_http(ctx, spkUrl, dir)
	default:
		err = fmt.Errorf("spk url is not support: %s", spkUrl)
	}
	return
}

func download_file(ctx context.Context, spkUrl string, dir string) (err error) {
	slog.InfoContext(ctx, "download spk file", "url", spkUrl)
	spkUrl = strings.TrimSuffix(spkUrl, "file://")
	f, e := os.Open(spkUrl)
	if err = e; err != nil {
		return
	}
	defer f.Close()

	err = Extract(ctx, f, dir, true)
	return
}

func download_http(ctx context.Context, spkUrl string, dir string) (err error) {
	slog.InfoContext(ctx, "download spk file", "url", spkUrl)
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, spkUrl, nil); err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5")

	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	err = Extract(ctx, resp.Body, dir, true)
	return
}

func allExists(ctx context.Context, dir string) bool {
	files := []string{
		filepath.Join(dir, "bin/bin/version"),
		filepath.Join(dir, "bin/bin/xunlei-pan-cli-launcher.{arch}"),
		filepath.Join(dir, "bin/bin/xunlei-pan-cli.{version}.{arch}"),
		filepath.Join(dir, "ui/index.cgi"),
	}

	version := utils.Cat(files[0])
	if version == "" {
		slog.DebugContext(ctx, "check spk fail, version not found")
		return false
	}
	slog.DebugContext(ctx, "check spk", "version", version)

	repl := strings.NewReplacer("{arch}", runtime.GOARCH, "{version}", version)
	for _, f := range files[1:] {
		f = repl.Replace(f)
		stat, err := os.Stat(f)
		if err != nil || !stat.Mode().IsRegular() {
			slog.DebugContext(ctx, "check spk fail", "file", f, "err", err)
			return false
		}
		slog.DebugContext(ctx, "check spk", "size", utils.HumanBytes(stat.Size()), "modtime", stat.ModTime(), "file", f)
	}
	return true
}
