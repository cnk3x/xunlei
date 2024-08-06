package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/cnk3x/xunlei/pkg/lod"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router = chi.Router

// UseBasicAuth implements a simple middleware handler for adding basic http auth to a route.
func UseBasicAuth(r Router, username, password string) {
	if password != "" {
		if username == "" {
			username = "admin"
		}
		r.Use(middleware.BasicAuth("xlp", map[string]string{username: password}))
	}
}

func NewMux() Router {
	router := chi.NewMux()
	router.Use(middleware.Recoverer)
	return router
}

type ServeOption struct {
	Handler    http.Handler
	Addr       string
	OnShutDown func(ctx context.Context)
}

func Serve(ctx context.Context, options ServeOption) (port uint16, err error) {
	portc := make(chan uint16)
	errc := make(chan error)

	s := &http.Server{
		Addr:    options.Addr,
		Handler: options.Handler,
		BaseContext: func(l net.Listener) context.Context {
			addr, _ := l.Addr().(*net.TCPAddr)
			if addr != nil {
				trySend(portc, uint16(addr.Port))
			}
			slog.InfoContext(ctx, "web started", "listen", l.Addr().String())
			return ctx
		},
	}

	onShutDown := func() {
		if options.OnShutDown != nil {
			options.OnShutDown(ctx)
		}
	}

	done := make(chan struct{})

	go func() {
		select {
		case <-done:
			return
		case <-ctx.Done():
			_ = s.Shutdown(context.Background())
		}
	}()

	go func() {
		defer close(done)
		defer close(portc)
		defer onShutDown()

		trySend(errc, s.ListenAndServe())
		if err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "web is done!", "err", err)
			return
		}
		slog.InfoContext(ctx, "web is done!")
	}()

	select {
	case port = <-portc:
		close(errc)
		err = <-errc
	case err = <-errc:
		close(errc)
	}
	return
}

func trySend[T any](c chan T, v T) {
	select {
	case <-c:
	default:
		c <- v
	}
}

func Blob[T ~[]byte | ~string](body T, contentType string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func Redirect(to string, permanent ...bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, lod.Iif(lod.Selects(permanent), http.StatusPermanentRedirect, http.StatusTemporaryRedirect))
	}
}

func JSON(value any, status int) http.HandlerFunc {
	data, err := json.Marshal(value)
	if err != nil {
		return Blob(fmt.Sprintf(`{"code": 500, "err": %s}`, strconv.Quote(err.Error())), "application/json", 500)
	}
	return Blob(data, "application/json", status)
}
