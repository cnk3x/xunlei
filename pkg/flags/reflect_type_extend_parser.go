package flags

import (
	"fmt"
	"net"
	"net/netip"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

func init() {
	Extend(rParseTime, rFormatTime, "time")
	Extend(rParseIp, rFormatIp, "ip")
	Extend(rParseDuration, rFormatDuration, "duration")
	Extend(rParsePathList, rFormatPathList, "paths")
}

func rParseTime(_ time.Time, s string) (t time.Time, err error) {
	if s != "" {
		var allowTimeLayouts = []string{
			time.DateTime, "2006-01-02 15:04", time.DateOnly, "01-02",
			"2006/01/02 15:04:05", "2006/01/02 15:04", "2006/01/02", "01/02",
			time.TimeOnly,
		}

		if t, err = time.Parse(time.RFC3339, s); err == nil {
			return
		}

		for _, layout := range allowTimeLayouts {
			if t, err = time.ParseInLocation(layout, s, time.Local); err == nil {
				return
			}
		}
	}
	return
}

func rFormatTime(in time.Time) (s string) {
	if !in.IsZero() {
		s = strings.TrimSuffix(strings.TrimSuffix(in.Format(time.DateTime), ":00"), " 00:00")
	}
	return
}

func rParseDuration(_ time.Duration, in string) (d time.Duration, err error) {
	if in != "" {
		if days, hours, found := strings.Cut(in, "d"); found {
			if hours != "" {
				if d, err = time.ParseDuration(hours); err != nil {
					return
				}
			}

			if days != "" {
				if _d, e := strconv.ParseInt(days, 10, 64); e == nil {
					const h24 = 24 * time.Hour
					d += time.Duration(_d) * h24
				} else {
					err = e
				}
			}
		} else {
			d, err = time.ParseDuration(in)
		}
	}
	return
}

func rFormatDuration(in time.Duration) (s string) {
	if in != 0 {
		const h24 = 24 * time.Hour
		dRepl := strings.NewReplacer("0d", "", "0h", "", "0m", "", "0s", "")
		s = dRepl.Replace(strconv.FormatInt(int64(in/h24), 10) + "d" + (in % h24).String())
	}
	return
}

func rParseIp(_ net.IP, s string) (r net.IP, err error) {
	if s != "" {
		if ip, e := netip.ParseAddr(s); e != nil {
			err = e
		} else if z := ip.Zone(); z != "" {
			err = fmt.Errorf("ipv6 zone is not supported: %s", z)
		} else {
			n := ip.As16()
			r = net.IP(n[:])
		}
	}
	return
}

func rFormatIp(in net.IP) (s string) {
	if len(in) > 0 {
		if s = in.String(); strings.Contains(s, "invalid") {
			s = ""
		}
	}
	return
}

type PathList []string

func (in PathList) Compact() (out PathList) {
	for _, p := range in {
		if p != "" && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	return
}

func (in PathList) String() string {
	return strings.Join(in.Compact(), string(filepath.ListSeparator))
}

func (in PathList) Abs() (out PathList) {
	for _, p := range in {
		if p != "" {
			if p, e := filepath.Abs(p); e == nil && !slices.Contains(out, p) {
				out = append(out, p)
			}
		}
	}
	return
}

func (in PathList) IfNil(s ...string) PathList {
	if len(in) == 0 {
		return append(in, s...)
	}
	return in
}

func (in PathList) Len() int {
	return len(in)
}

func rParsePathList(src PathList, s string) (r PathList, err error) {
	if s != "" {
		r = append(src, filepath.SplitList(s)...)
	}
	return
}

func rFormatPathList(in PathList) (s string) {
	if len(in) > 0 {
		s = strings.Join(in, string(filepath.ListSeparator))
	}
	return
}
