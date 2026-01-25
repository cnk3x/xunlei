package spk

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/utils"
	"github.com/cnk3x/xunlei/pkg/vms/sys"
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
	spkUrl = strings.TrimPrefix(spkUrl, "file://")
	slog.InfoContext(ctx, "download spk file", "url", spkUrl)
	f, e := os.Open(spkUrl)
	if err = e; err != nil {
		return
	}
	defer f.Close()

	err = Extract(ctx, f, dir)
	return
}

func download_http(ctx context.Context, spkUrl string, dir string) (err error) {
	slog.InfoContext(ctx, "download spk file", "url", spkUrl)
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, spkUrl, nil); err != nil {
		return
	}
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-encoding", "gzip, deflate, br, zstd")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,zh-TW;q=0.5")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("dnt", "1")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=0, i")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0")

	cli := &http.Client{
		Timeout: time.Second * 60,
		Transport: &http.Transport{
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 10 * time.Second}).DialContext,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}
	var resp *http.Response
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	err = Extract(ctx, resp.Body, dir)
	return
}

func allExists(ctx context.Context, dir string) bool {
	files := []string{
		filepath.Join(dir, "bin/bin/version"),
		filepath.Join(dir, "bin/bin/xunlei-pan-cli-launcher.{arch}"),
		filepath.Join(dir, "bin/bin/xunlei-pan-cli.{version}.{arch}"),
		filepath.Join(dir, "ui/index.cgi"),
	}

	version := fo.Cat(files[0], true)
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
		slog.DebugContext(ctx, "check spk", "perm", sys.Perm2s(stat.Mode()), "size", utils.HumanBytes(stat.Size()), "modtime", stat.ModTime(), "file", f)
	}
	return true
}
