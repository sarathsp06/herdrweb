// Command herdr-bridge serves the Herdr Web UI and proxies the Herdr socket.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarathsp06/herdrweb/internal/config"
	"github.com/sarathsp06/herdrweb/internal/herdr"
	"github.com/sarathsp06/herdrweb/internal/server"
)

// version is overridden at release time via -ldflags.
var version = "dev"

func main() {
	addr := flag.String("addr", "127.0.0.1:7331", "listen address (loopback only by default)")
	socket := flag.String("socket", herdr.DefaultSocketPath(), "path to the Herdr socket")
	cfgPath := flag.String("config", config.DefaultPath(), "path to Herdr config.toml")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		log.SetFlags(0)
		log.Println(version)
		return
	}

	client := herdr.New(*socket)
	hub := server.NewHub(client, *cfgPath, version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go hub.Run(ctx)

	handler, err := hub.Handler()
	if err != nil {
		log.Fatalf("build handler: %v", err)
	}
	srv := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()

	log.Printf("herdr-bridge %s listening on http://%s (socket %s)", version, *addr, *socket)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}
