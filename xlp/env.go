package xlp

import (
	"os"
	"slices"
	"strings"
)

type Env struct {
	envMap map[string]string
	keys   []string
}

func (e Env) init() *Env {
	e.envMap = make(map[string]string)
	return &e
}

func (e *Env) Each(walkFn func(k, v string)) {
	for _, k := range e.keys {
		walkFn(k, e.envMap[k])
	}
}

func (e *Env) Set(k, v string) *Env {
	if k = strings.TrimSpace(k); k != "" {
		idx := slices.Index(e.keys, k)
		if v == "" {
			delete(e.envMap, k)
			if idx > -1 {
				e.keys = slices.Delete(e.keys, idx, idx+1)
			}
		} else {
			e.envMap[k] = v
			if idx == -1 {
				e.keys = append(e.keys, k)
			}
		}
	}

	return e
}

func (e *Env) Append(envs ...string) *Env {
	for _, it := range envs {
		k, v, found := strings.Cut(it, "=")
		if !found {
			k = it
		}
		e.Set(k, v)
	}
	return e
}

func (e *Env) WithOS() *Env {
	return e.Append(os.Environ()...)
}

func (e *Env) Environ() []string {
	return Map(e.keys, func(k string) string { return k + "=" + e.envMap[k] })
}
