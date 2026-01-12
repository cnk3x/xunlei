package xunlei

import (
	"cmp"
	"fmt"
	"net/http"

	"github.com/cnk3x/xunlei/pkg/utils"
)

func webRedirect(to string, permanent ...bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, utils.Iif(cmp.Or(permanent...), http.StatusPermanentRedirect, http.StatusTemporaryRedirect))
	}
}

func webBlob[T ~[]byte | ~string](body T, contentType string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

type webMw = func(next http.Handler) http.Handler

func webRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}
				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// func webChain(h http.Handler, mws ...webMw) http.Handler {
// 	for _, mw := range slices.Backward(mws) {
// 		h = mw(h)
// 	}
// 	return h
// }

// func webHandle(mux *http.ServeMux, pattern string, h http.Handler, mws ...webMw) {
// 	mux.Handle(pattern, webChain(h, slices.Insert(mws, 0, webRecoverer)...))
// }

func webBasicAuth(username, password string) webMw {
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
