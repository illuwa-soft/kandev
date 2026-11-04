// Package weburl validates operator-configured public web origins without I/O.
package weburl

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var ErrOrigin = errors.New("public web URL must be an HTTPS origin (HTTP allowed only on loopback), without credentials, path, query or fragment")

func ParseOrigin(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.Opaque != "" || u.User != nil || strings.ContainsAny(raw, "?#\r\n\t ") {
		return nil, ErrOrigin
	}
	if u.Path != "" && u.Path != "/" || u.RawPath != "" {
		return nil, ErrOrigin
	}
	if !allowedScheme(u) || !validPort(u) {
		return nil, ErrOrigin
	}
	u.Path = ""
	return u, nil
}

func allowedScheme(u *url.URL) bool {
	ip := net.ParseIP(u.Hostname())
	loopback := strings.EqualFold(u.Hostname(), "localhost") || ip != nil && ip.IsLoopback()
	return u.Scheme == "https" || u.Scheme == "http" && loopback
}

func validPort(u *url.URL) bool {
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return false
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return false
	}
	return true
}
