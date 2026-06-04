package config

import "testing"

// All tests use t.Setenv so values auto-reset between tests. Setting to "" is
// equivalent to unset for our envOr semantics (which only consults non-empty values).

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"UNIC_BIND", "UNIC_DB", "UNIC_ADMIN_PASSWORD",
		"UNIC_SSH_HOST", "UNIC_SSH_PORT", "UNIC_PROVISIONER", "UNIC_SUDO_SCRIPT",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadRequiresHost(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_ADMIN_PASSWORD", "x")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when SSH_HOST is empty")
	}
}

func TestLoadRequiresAdminPassword(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_SSH_HOST", "1.2.3.4")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when ADMIN_PASSWORD is empty")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_SSH_HOST", "1.2.3.4")
	t.Setenv("UNIC_ADMIN_PASSWORD", "secret")
	c, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if c.BindAddr != "127.0.0.1:8080" {
		t.Fatalf("BindAddr: %q", c.BindAddr)
	}
	if c.DBPath != "unic-panel.db" {
		t.Fatalf("DBPath: %q", c.DBPath)
	}
	if c.SSHPort != 22 {
		t.Fatalf("SSHPort: %d", c.SSHPort)
	}
	if c.Provisioner != "stub" {
		t.Fatalf("Provisioner: %q", c.Provisioner)
	}
	if c.SudoScriptPath != "/usr/local/sbin/tunnel-user.sh" {
		t.Fatalf("SudoScriptPath: %q", c.SudoScriptPath)
	}
}

func TestLoadHonoursOverrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_SSH_HOST", "198.105.115.89")
	t.Setenv("UNIC_SSH_PORT", "2222")
	t.Setenv("UNIC_ADMIN_PASSWORD", "secret")
	t.Setenv("UNIC_PROVISIONER", "sudo")
	t.Setenv("UNIC_BIND", "0.0.0.0:9000")
	c, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if c.SSHHost != "198.105.115.89" || c.SSHPort != 2222 ||
		c.Provisioner != "sudo" || c.BindAddr != "0.0.0.0:9000" {
		t.Fatalf("overrides not honoured: %+v", c)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_SSH_HOST", "1.2.3.4")
	t.Setenv("UNIC_ADMIN_PASSWORD", "x")
	t.Setenv("UNIC_SSH_PORT", "not-a-port")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

func TestLoadOutOfRangePort(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_SSH_HOST", "1.2.3.4")
	t.Setenv("UNIC_ADMIN_PASSWORD", "x")
	t.Setenv("UNIC_SSH_PORT", "99999")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for out-of-range port")
	}
}

func TestLoadInvalidProvisioner(t *testing.T) {
	clearEnv(t)
	t.Setenv("UNIC_SSH_HOST", "1.2.3.4")
	t.Setenv("UNIC_ADMIN_PASSWORD", "x")
	t.Setenv("UNIC_PROVISIONER", "garbage")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid provisioner")
	}
}
