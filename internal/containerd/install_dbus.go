package containerd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// UsesSystemd checks if the system is using systemd
// by running the "systemctl is-system-running" command.
// NOTE: this limits support to systems using systemctl to manage systemd
func UsesSystemd() bool {
	// Check if is a systemd system
	cmd := nsenterCmd("systemctl", "is-system-running", "--quiet")
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
	whichApt := nsenterCmd("which", "apt-get")
	whichYum := nsenterCmd("which", "yum")
	whichDnf := nsenterCmd("which", "dnf")
	whichApk := nsenterCmd("which", "apk")
	if err := whichApt.Run(); err == nil {
		cmd = nsenterCmd("apt-get", "update", "--yes")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update apt: %w", err)
		}
		cmd = nsenterCmd("apt-get", "install", "--yes", "dbus")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install D-Bus with apt: %w", err)
		}
	} else if err = whichDnf.Run(); err == nil {
		cmd = nsenterCmd("dnf", "install", "--yes", "dbus")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install D-Bus with dnf: %w", err)
		}
	} else if err = whichApk.Run(); err == nil {
		cmd = nsenterCmd("apk", "add", "dbus")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install D-Bus with apk: %w", err)
		}
	} else if err = whichYum.Run(); err == nil {
		cmd = nsenterCmd("yum", "install", "--yes", "dbus")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install D-Bus with yum: %w", err)
		}
	} else {
		slog.Info("WARNING: Could not install D-Bus. No supported package manager found.")
		return nil
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
