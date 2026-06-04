// Package config holds the typed runtime config loaded from env vars.
package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	BindAddr       string // UNIC_BIND (default 127.0.0.1:8080)
	DBPath         string // UNIC_DB (default unic-panel.db)
	AdminPassword  string // UNIC_ADMIN_PASSWORD (REQUIRED on first run; sets admin password)
	SSHHost        string // UNIC_SSH_HOST (REQUIRED; e.g. 198.105.115.89)
	SSHPort        int    // UNIC_SSH_PORT (default 22; this VPS uses 2222)
	Provisioner    string // UNIC_PROVISIONER: "stub" (dev) or "sudo" (VPS). Default "stub".
	SudoScriptPath string // UNIC_SUDO_SCRIPT (default /usr/local/sbin/tunnel-user.sh)
}

func Load() (*Config, error) {
	c := &Config{
		BindAddr:       envOr("UNIC_BIND", "127.0.0.1:8080"),
		DBPath:         envOr("UNIC_DB", "unic-panel.db"),
		AdminPassword:  os.Getenv("UNIC_ADMIN_PASSWORD"),
		SSHHost:        os.Getenv("UNIC_SSH_HOST"),
		Provisioner:    envOr("UNIC_PROVISIONER", "stub"),
		SudoScriptPath: envOr("UNIC_SUDO_SCRIPT", "/usr/local/sbin/tunnel-user.sh"),
	}

	port, err := strconv.Atoi(envOr("UNIC_SSH_PORT", "22"))
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("UNIC_SSH_PORT must be a valid port number 1-65535")
	}
	c.SSHPort = port

	if c.SSHHost == "" {
		return nil, errors.New("UNIC_SSH_HOST is required (e.g. 198.105.115.89)")
	}
	if c.AdminPassword == "" {
		return nil, errors.New("UNIC_ADMIN_PASSWORD is required on first run")
	}
	if c.Provisioner != "stub" && c.Provisioner != "sudo" {
		return nil, errors.New(`UNIC_PROVISIONER must be "stub" or "sudo"`)
	}
	return c, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
