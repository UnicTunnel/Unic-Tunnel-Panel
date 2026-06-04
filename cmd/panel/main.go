package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/accounts"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/auth"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/config"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/server"
	"github.com/UnicTunnel/Unic-Tunnel-Panel/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	// EnsureAdmin upserts, so updating UNIC_ADMIN_PASSWORD and restarting rotates it.
	hash, err := auth.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Fatalf("hash admin pw: %v", err)
	}
	if err := st.EnsureAdmin(context.Background(), "admin", hash); err != nil {
		log.Fatalf("ensure admin: %v", err)
	}

	var prov accounts.Provisioner
	switch cfg.Provisioner {
	case "stub":
		prov = accounts.NewStub()
	case "sudo":
		prov = accounts.NewSudo(cfg.SudoScriptPath)
	}

	srv := server.New(server.Deps{
		Store:       st,
		Provisioner: prov,
		SSHHost:     cfg.SSHHost,
		SSHPort:     cfg.SSHPort,
	})

	httpSrv := &http.Server{
		Addr:              cfg.BindAddr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("unic-panel listening on http://%s (provisioner=%s, ssh=%s:%d)",
			cfg.BindAddr, cfg.Provisioner, cfg.SSHHost, cfg.SSHPort)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
}
