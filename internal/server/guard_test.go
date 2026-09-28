package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHostGuard pins who may reach the bridge. The bridge has no auth and
// forwards input to live terminals, so a request a browser sent on an
// attacker's behalf (DNS rebinding, cross-site WebSocket) must never pass,
// while every legitimate route (loopback, tailnet IP, `tailscale serve`,
// Vite dev proxy, explicit -allow-host) must.
func TestHostGuard(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	g := newHostGuard(ok, []string{"Herdr.Example.com."})

	cases := []struct {
		name, host, origin string
		want               int
	}{
		{"loopback, no origin (curl, native clients)", "127.0.0.1:7331", "", http.StatusNoContent},
		{"loopback same-origin page", "127.0.0.1:7331", "http://127.0.0.1:7331", http.StatusNoContent},
		{"localhost same-origin", "localhost:7331", "http://localhost:7331", http.StatusNoContent},
		{"ipv6 loopback", "[::1]:7331", "http://[::1]:7331", http.StatusNoContent},
		{"bound to tailnet IP", "100.64.0.7:7331", "http://100.64.0.7:7331", http.StatusNoContent},
		{"tailscale serve (Host preserved)", "desktop.tail1234.ts.net", "https://desktop.tail1234.ts.net", http.StatusNoContent},
		{"vite dev proxy rewrites Host, keeps loopback Origin", "127.0.0.1:7331", "http://localhost:5173", http.StatusNoContent},
		{"-allow-host is case- and dot-insensitive", "HERDR.example.com:443", "https://HERDR.example.com:443", http.StatusNoContent},

		{"DNS rebinding: attacker name resolved to loopback", "evil.example:7331", "http://evil.example:7331", http.StatusForbidden},
		{"rebinding via IP-looking name", "127.0.0.1.nip.io:7331", "", http.StatusForbidden},
		{"cross-site page hits loopback", "127.0.0.1:7331", "https://evil.example", http.StatusForbidden},
		{"cross-site page served from a public IP", "127.0.0.1:7331", "http://203.0.113.9", http.StatusForbidden},
		{"another tailnet's funnel page", "desktop.tail1234.ts.net", "https://evil.tail9999.ts.net", http.StatusForbidden},
		{"same host, different port is another origin", "100.64.0.7:7331", "http://100.64.0.7:8080", http.StatusForbidden},
		{"opaque origin (sandboxed iframe)", "127.0.0.1:7331", "null", http.StatusForbidden},
		{"empty Host", "", "", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ws", nil)
			r.Host = tc.host
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			w := httptest.NewRecorder()
			g.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("host=%q origin=%q: status %d, want %d", tc.host, tc.origin, w.Code, tc.want)
			}
		})
	}
}
