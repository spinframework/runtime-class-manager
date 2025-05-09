package containerd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// UsesSystemd checks if the system is using systemd.
func UsesSystemd() bool {
	cmd := nsenterCmd("systemctl", "list-units", "|", "grep", "-q", "containerd.service")
	if err := cmd.Run(); err != nil {
		slog.Info("Error with systemctl: %w\n", "error", err)
		return false
	}
	return true
}

// InstallDbus checks if D-Bus service is installed and active. If not, installs D-Bus
// and starts the service.
// NOTE: this limits support to systems using systemctl to manage systemd.
func InstallDbus() error {
	cmd := nsenterCmd("systemctl", "is-active", "dbus", "--quiet")
	if err := cmd.Run(); err == nil {
		slog.Info("D-Bus is already installed and running")
		return nil
	}
	slog.Info("installing D-Bus")

	type pkgManager struct {
		name    string
		check   []string
		install []string
	}

	managers := []pkgManager{
		{"apt-get", []string{"which", "apt-get"}, []string{"apt-get", "update", "--yes", "&&", "apt-get", "install", "--yes", "dbus"}},
		{"dnf", []string{"which", "dnf"}, []string{"dnf", "install", "--yes", "dbus"}},
		{"apk", []string{"which", "apk"}, []string{"apk", "add", "dbus"}},
		{"yum", []string{"which", "yum"}, []string{"yum", "install", "--yes", "dbus"}},
	}
	installed := false
	for _, mgr := range managers {
		if err := nsenterCmd(mgr.check...).Run(); err == nil {
			if err := nsenterCmd(mgr.install...).Run(); err != nil {
				return fmt.Errorf("failed to install D-Bus with %s: %w", mgr.name, err)
			}
			installed = true
			break
		}
	}

	if !installed {
		return fmt.Errorf("could not install D-Bus as no supported package manager found")
	}

	slog.Info("restarting D-Bus")
	cmd = nsenterCmd("systemctl", "restart", "dbus")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart D-Bus: %w", err)
	}

	return nil
}

func nsenterCmd(cmd ...string) *exec.Cmd {
	return exec.Command("nsenter",
		append([]string{fmt.Sprintf("-m/%s/proc/1/ns/mnt", os.Getenv("HOST_ROOT")), "--"}, cmd...)...) // #nosec G204
}
