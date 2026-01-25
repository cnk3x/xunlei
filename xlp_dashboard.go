package xunlei

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/cgi"
	"strconv"
	"strings"
	"time"

	"github.com/cnk3x/xunlei/pkg/cmdx"
	"github.com/cnk3x/xunlei/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func dashboardRun(ctx context.Context, cfg *Config) (done <-chan struct{}, err error) {
	addr := cfg.DashboardIp.String()
	if addr == "<nil>" {
		addr = ""
	}
	addr = net.JoinHostPort(addr, strconv.FormatUint(uint64(cfg.DashboardPort), 10))
	return wsRun(ctx, addr, dashboardRouter(ctx, cfg), "dashboard")
}

func dashboardRouter(ctx context.Context, cfg *Config) http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Recoverer)

	d := make([]byte, (13+4)/8*5)
	rand.Read(d)
	synoToken := base32.StdEncoding.EncodeToString(d)
	if len(synoToken) > 13 {
		synoToken = synoToken[:13]
	}

	token := fmt.Appendf(make([]byte, 0, 64), `{"SynoToken":%q,"result":"success","success":true}`, synoToken)
	mux.Handle("/webman/login.cgi", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.Write(token)
	}))

	lErr := log.Line(func(s string) { slog.DebugContext(ctx, "[cgi] [err] "+s) })
	defer lErr.Close()

	lLog := cmdx.LineWriter(func(s string) { slog.DebugContext(ctx, "[cgi] [log] "+s) })
	defer lLog.Close()

	const CGI_PATH = "/webman/3rdparty/" + SYNOPKG_PKGNAME + "/index.cgi/"
	mux.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("xunlei", map[string]string{cfg.DashboardUsername: cfg.DashboardPassword}))
		r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, CGI_PATH, 308) }))
		r.Mount(CGI_PATH, &cgi.Handler{
			Path:   FILE_INDEX_CGI,
			Dir:    DIR_SYNOPKG_WORK,
			Env:    mockEnv(cfg.DirData, strings.Join(cfg.DirDownload, ":")),
			Stderr: lErr,
			Logger: log.Std(lLog),
		})
	})

	return mux
}

func wsRun(ctx context.Context, addr string, handler http.Handler, name string) (done <-chan struct{}, err error) {
	s := &http.Server{Addr: addr, Handler: handler}

	cDone := make(chan struct{})
	done = cDone

	cStarted := make(chan error, 1)
	s.BaseContext = func(ln net.Listener) context.Context {
		slog.InfoContext(ctx, name+" started", "listen", ln.Addr().String())
		return ctx
	}

	go func() {
		defer close(cStarted)
		defer close(cDone)

		ln, err := net.Listen("tcp", addr)
		cStarted <- err

		if err == nil {
			err = s.Serve(ln)
		}

		if err != nil && err != http.ErrServerClosed {
			slog.WarnContext(ctx, name+" done", "err", err.Error())
		} else {
			slog.InfoContext(ctx, name+" done")
		}
	}()

	if err = <-cStarted; err == nil {
		go func() {
			<-ctx.Done()
			s.Shutdown(context.Background())
		}()
	}

	select {
	case <-done:
	case <-time.After(time.Millisecond * 100):
	}
	return
}
