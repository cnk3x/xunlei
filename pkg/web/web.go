// Package web provides a simple HTTP multiplexer with middleware support.
package web

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/cnk3x/xunlei/pkg/utils"
)

// Processor 定义中间件处理器函数类型，接收下一个处理器并返回包装后的处理器
type Processor func(next http.Handler) http.Handler

// Mux 表示一个HTTP请求多路复用器，支持中间件处理链
type Mux struct {
	cur        *http.ServeMux // 当前多路复用器
	parent     *Mux           // 父级多路复用器，用于嵌套结构
	processors []Processor    // 处理器链（中间件）
	onShutdown func()         // 关闭时执行的回调函数
}

// Router 接口定义路由的基本操作
type Router interface {
	Handle(pattern string, handler http.Handler)
}

// NewMux 创建一个新的Mux实例，默认包含Recoverer中间件
func NewMux() *Mux {
	return &Mux{cur: http.NewServeMux()}
}

// OnShutDown 设置服务器关闭时的回调函数
//   - fn: 服务器关闭时执行的函数
func (mux *Mux) OnShutDown(fn func()) { mux.onShutdown = fn }

// With 创建一个新的Mux实例，继承当前实例的父级和处理器，并可以添加新的处理器
//   - processor: 要添加的处理器列表
//
// 返回一个新的Mux实例，其父级指向当前实例
func (mux *Mux) With(processor ...Processor) *Mux { return &Mux{parent: mux} }

func (mux *Mux) BasicAuth(user, pwd string) *Mux { return mux.With(BasicAuth(user, pwd)) }

// Use 向当前mux添加中间件处理器到处理器链中
//   - processor: 要添加的处理器列表
func (mux *Mux) Use(processor ...Processor) { mux.processors = append(mux.processors, processor...) }

func (mux *Mux) UseBasicAuth(user, pwd string) { mux.Use(BasicAuth(user, pwd)) }

func (mux *Mux) UseRecoverer() { mux.Use(Recoverer) }

// Handle 注册HTTP处理器，根据是否有父级mux来决定注册位置
//   - pattern: 请求路径模式
//   - handler: HTTP处理器
//   - processors: 额外的处理器（中间件）
func (mux *Mux) Handle(pattern string, handler http.Handler) {
	slog.Debug("web handle", "pattern", pattern, "handler", handler != nil, "parent", mux.parent != nil, "cur", mux.cur != nil)
	if mux.parent != nil {
		mux.parent.Handle(pattern, applyProcessors(handler, mux.processors...))
	} else {
		mux.cur.Handle(pattern, applyProcessors(handler, mux.processors...))
	}
}

// Route 注册一个带有前缀模式和处理器的路由，允许通配符匹配。
// 如果前缀不以'*'结尾，则根据情况添加'*'或'/*'以实现路径前缀匹配。
//   - prefix: 要匹配的路由前缀（可选尾随'*'）
//   - handler: 路由匹配时执行的HTTP处理器
func (mux *Mux) Route(prefix string, handler http.Handler) {
	if !strings.HasSuffix(prefix, "*") {
		if strings.HasSuffix(prefix, "/") {
			prefix += "*"
		} else {
			prefix += "/*"
		}
	}
	mux.Handle(prefix, handler)
}

// Get 注册GET请求处理器
//   - pattern: GET请求路径模式
//   - handler: HTTP处理器
//   - processors: 额外的处理器（中间件）
func (mux *Mux) Get(pattern string, handler http.Handler) {
	mux.Handle("GET "+pattern, handler)
}

// Post 注册POST请求处理器
//   - pattern: POST请求路径模式
//   - handler: HTTP处理器
//   - processors: 额外的处理器（中间件）
func (mux *Mux) Post(pattern string, handler http.Handler) {
	mux.Handle("POST "+pattern, handler)
}

// ServeHTTP 实现http.Handler接口，处理HTTP请求
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mux.parent != nil {
		mux.parent.ServeHTTP(w, r)
	} else if mux.cur != nil {
		mux.cur.ServeHTTP(w, r)
	} else {
		http.Error(w, "mux is not defined", http.StatusInternalServerError)
	}
}

// Run 启动HTTP服务器并监听指定地址
//   - ctx: 上下文对象
//   - addr: 监听地址
//
// 返回启动错误（如果有的话）
func (mux *Mux) Run(ctx context.Context, addr string) (err error) {
	s := &http.Server{
		Addr: addr, Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			slog.InfoContext(ctx, "web started", "listen", l.Addr().String())
			return ctx
		},
	}

	if err = s.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	if mux.onShutdown != nil {
		mux.onShutdown()
	}
	return
}

// Start 在后台启动HTTP服务器
//   - ctx: 上下文对象
//   - addr: 监听地址
//
// 返回完成通道，当服务器停止时该通道会被关闭
func (mux *Mux) Start(ctx context.Context, addr string) (done <-chan struct{}) {
	webDone := make(chan struct{})
	go func() {
		defer close(webDone)
		if err := mux.Run(ctx, addr); err != nil {
			slog.ErrorContext(ctx, "web is done!", "err", err)
		} else {
			slog.InfoContext(ctx, "web is done!")
		}
	}()
	return webDone
}

// Redirect 创建重定向处理器
//   - to: 重定向目标URL
//   - permanent: 是否永久重定向，默认临时重定向
//
// 返回HTTP处理器函数
func Redirect(to string, permanent ...bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, utils.Iif(cmp.Or(permanent...), http.StatusPermanentRedirect, http.StatusTemporaryRedirect))
	}
}

// Blob 创建返回字节切片或字符串内容的处理器
//   - body: 响应体内容
//   - contentType: 内容类型
//   - status: HTTP状态码
//
// 返回HTTP处理器函数
func Blob[T ~[]byte | ~string](body T, contentType string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func JSON(body any, status int) http.HandlerFunc {
	return Blob(utils.Eon(json.Marshal(body)), "application/json", status)
}

func FBlob[T ~[]byte | ~string](body func() (T, error), contentType string) http.HandlerFunc {
	headerContentType := "Content-Type"
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := body()
		if err == nil {
			w.Header().Set(headerContentType, contentType)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(data))
		} else {
			w.Header().Set(headerContentType, "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}
}

func FJSON[T any](body func() (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := body()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		data, err := json.Marshal(val)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(data)
	}
}

// Recoverer 中间件，从panic中恢复并记录日志
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					slog.DebugContext(r.Context(), "abort")
					// panic(rvr)
				}
				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// BasicAuth 创建基础认证中间件
//   - username: 用户名
//   - password: 密码
//
// 返回认证处理器
func BasicAuth(username, password string) Processor {
	if password != "" {
		if username == "" {
			username = "admin"
		}
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				if !ok || user != username || pass != password {
					w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "xlp"))
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
			})
		}
	}
	return func(next http.Handler) http.Handler { return next }
}

// Address 组合IP地址和端口为网络地址字符串
//   - ip: IP地址
//   - port: 端口号
//
// 返回组合的网络地址字符串
func Address[T utils.UintT | utils.IntT](ip net.IP, port T) string {
	sIp := ip.String()
	return net.JoinHostPort(utils.Iif(sIp == "<nil>", "", sIp), utils.String(port))
}

// applyProcessors 按逆序应用处理器链
//   - h: 初始处理器
//   - mw: 处理器列表
//
// 返回包装后的处理器
func applyProcessors(h http.Handler, mw ...Processor) http.Handler {
	for _, m := range slices.Backward(mw) {
		if m != nil {
			h = m(h)
		}
	}
	return h
}
