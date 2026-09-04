// Command herdr-bridge serves the Herdr Web UI and proxies the Herdr socket.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sarathsp06/herdrweb/internal/config"
	"github.com/sarathsp06/herdrweb/internal/daemon"
	"github.com/sarathsp06/herdrweb/internal/herdr"
	"github.com/sarathsp06/herdrweb/internal/push"
	"github.com/sarathsp06/herdrweb/internal/server"
	"github.com/sarathsp06/herdrweb/internal/service"
)

// version is overridden at release time via -ldflags.
var version = "dev"

func main() {
	addr := flag.String("addr", "127.0.0.1:7331", "listen address (loopback only by default)")
	socket := flag.String("socket", herdr.DefaultSocketPath(), "path to the Herdr socket")
	cfgPath := flag.String("config", config.DefaultPath(), "path to Herdr config.toml")
	logPath := flag.String("log-file", "", "path to write application logs")
	pidPath := flag.String("pid-file", "", "path to write process PID file")
	serviceAction := flag.String("service", "", "manage system service: install, uninstall, start, stop, status")

	var isDaemonLong, isDaemonShort bool
	flag.BoolVar(&isDaemonLong, "daemon", false, "run in background as a daemon")
	flag.BoolVar(&isDaemonShort, "d", false, "run in background as a daemon (shorthand)")

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		log.SetFlags(0)
		log.Println(version)
		return
	}

	if *serviceAction != "" {
		opts := service.ServiceOptions{
			Addr:    *addr,
			Socket:  *socket,
			Config:  *cfgPath,
			LogPath: *logPath,
		}
		if err := service.Manage(*serviceAction, opts); err != nil {
			log.Fatalf("service %s: %v", *serviceAction, err)
		}
		return
	}

	if isDaemonLong || isDaemonShort {
		isParent, pid, err := daemon.Daemonize(*logPath)
		if err != nil {
			log.Fatalf("daemonize: %v", err)
		}
		if isParent {
			fmt.Printf("herdr-bridge daemon started (PID %d)\n", pid)
			return
		}
	}

	if *logPath != "" {
		logFile, err := daemon.SetupLogging(*logPath)
		if err != nil {
			log.Fatalf("setup logging: %v", err)
		}
		if logFile != nil {
			defer logFile.Close()
		}
	}

	if *pidPath != "" {
		cleanupPID, err := daemon.WritePIDFile(*pidPath)
		if err != nil {
			log.Fatalf("write pid file: %v", err)
		}
		defer cleanupPID()
	}

	client := herdr.New(*socket)
	pm, err := push.New(filepath.Dir(*cfgPath))
	if err != nil {
		log.Printf("web push disabled: %v", err)
	}
	hub := server.NewHub(client, *cfgPath, version, pm)

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
