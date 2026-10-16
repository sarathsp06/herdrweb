package server

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// hostGuard rejects requests a browser could only have sent on an attacker's
// behalf. The bridge has no auth and forwards input to live terminals, so:
//
//   - Host must be something DNS rebinding cannot forge: an IP literal,
//     localhost, a Tailscale MagicDNS name (*.ts.net, e.g. behind
//     `tailscale serve`), or a name passed via -allow-host.
//   - Origin, when present, must be the same host:port as Host, or loopback
//     (the Vite dev proxy rewrites Host but not Origin). This blocks
//     cross-site WebSocket hijacking from any page the operator happens to open.
type hostGuard struct {
	next  http.Handler
	extra map[string]bool
}

func newHostGuard(next http.Handler, allowHosts []string) *hostGuard {
	extra := map[string]bool{}
	for _, h := range allowHosts {
		if h = normalizeHost(h); h != "" {
			extra[h] = true
		}
	}
	return &hostGuard{next: next, extra: extra}
}

func (g *hostGuard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !g.allowedHost(r.Host) {
		log.Printf("rejected request for host %q (pass -allow-host %s to permit it)", r.Host, hostname(r.Host))
		http.Error(w, "host not allowed", http.StatusForbidden)
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && !allowedOrigin(origin, r.Host) {
		log.Printf("rejected cross-origin request from %q to %q", origin, r.Host)
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}
	g.next.ServeHTTP(w, r)
}

func (g *hostGuard) allowedHost(hostport string) bool {
	h := hostname(hostport)
	switch {
	case h == "":
		return false
	case net.ParseIP(h) != nil, h == "localhost", strings.HasSuffix(h, ".ts.net"):
		return true
	default:
		return g.extra[h]
	}
}

func allowedOrigin(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.EqualFold(u.Host, host) || isLoopback(normalizeHost(u.Hostname()))
}

func isLoopback(h string) bool {
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// hostname strips the port and normalizes a Host header value.
func hostname(hostport string) string {
	h := hostport
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		h = host
	}
	return normalizeHost(strings.Trim(h, "[]"))
}

func normalizeHost(h string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(h)), ".")
}
