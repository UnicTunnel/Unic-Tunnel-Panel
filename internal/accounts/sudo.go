package accounts

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// NewSudo returns a provisioner that calls a scoped sudoers script
// (default /usr/local/sbin/tunnel-user.sh) via `sudo -n` to manage tunnel
// users without giving the panel full root.
//
// The password is piped over stdin, NOT passed as a command-line argument,
// so it never appears in /proc/<pid>/cmdline or `ps` output.
//
// See scripts/sudoers.unicpanel for the whitelisting entry that makes this work.
func NewSudo(scriptPath string) Provisioner {
	return &sudoProv{script: scriptPath}
}

type sudoProv struct{ script string }

func (p *sudoProv) Create(ctx context.Context, username, password string) error {
	cmd := exec.CommandContext(ctx, "sudo", "-n", p.script, "create", username)
	cmd.Stdin = strings.NewReader(password)
	return p.run(cmd, "create", username)
}

func (p *sudoProv) Lock(ctx context.Context, username string) error {
	cmd := exec.CommandContext(ctx, "sudo", "-n", p.script, "lock", username)
	return p.run(cmd, "lock", username)
}

func (p *sudoProv) Delete(ctx context.Context, username string) error {
	cmd := exec.CommandContext(ctx, "sudo", "-n", p.script, "delete", username)
	return p.run(cmd, "delete", username)
}

func (p *sudoProv) run(cmd *exec.Cmd, action, username string) error {
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("sudo provisioner %s failed: user=%s err=%v out=%q",
			action, username, err, trimOutput(out))
		return fmt.Errorf("provisioner: %w", err)
	}
	return nil
}

func trimOutput(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		return s[:300] + "...(truncated)"
	}
	return s
}
