package containerd

import (
	"fmt"
	"log/slog"
)

// InstallDbus checks if D-Bus service is installed and active. If not, installs D-Bus
// and starts the service.
// NOTE: this limits support to systems using systemctl to manage systemd.
func InstallDbus() error {
	cmd := nsenterCmd("systemctl", "start", "dbus", "--quiet")
	if err := cmd.Run(); err == nil {
		slog.Info("D-Bus is already installed and running")
		return nil
	}
	slog.Info("installing D-Bus")

	type pkgManager struct {
		name    string
		check   []string
		update  []string
		install []string
	}

	managers := []pkgManager{
		{"apt-get", []string{"which", "apt-get"}, []string{"apt-get", "update", "--yes"}, []string{"apt-get", "install", "--yes", "dbus"}},
		{"dnf", []string{"which", "dnf"}, []string{}, []string{"dnf", "install", "--yes", "dbus"}},
		{"apk", []string{"which", "apk"}, []string{}, []string{"apk", "add", "dbus"}},
		{"yum", []string{"which", "yum"}, []string{}, []string{"yum", "install", "--yes", "dbus"}},
	}
	installed := false
	for _, mgr := range managers {
		if err := nsenterCmd(mgr.check...).Run(); err == nil {
			if len(mgr.update) != 0 {
				if err := nsenterCmd(mgr.update...).Run(); err != nil {
					return fmt.Errorf("failed to update package manager %s: %w", mgr.name, err)
				}
			}
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
