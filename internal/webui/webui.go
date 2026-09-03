// Package webui embeds the built SvelteKit assets so the bridge ships as a
// single binary. The Makefile copies web/build into ./dist before `go build`.
package webui

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"strings"
)

func init() {
	// Go's default MIME table lacks .webmanifest; without this the PWA manifest
	// is served as text/plain, which strict browsers may reject.
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

//go:embed all:dist
var embedded embed.FS

// Assets returns the embedded SvelteKit build rooted at dist/.
func Assets() (fs.FS, error) {
	return fs.Sub(embedded, "dist")
}

// Handler serves the SPA: static files when they exist, else index.html
// fallback (client-side routing).
func Handler() (http.Handler, error) {
	sub, err := Assets()
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(sub, p); err != nil {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL.Path = "/"
			http.ServeFileFS(w, r2, sub, "index.html")
			return
		}
		fileServer.ServeHTTP(w, r)
	}), nil
}
