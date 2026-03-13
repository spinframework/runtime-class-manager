//go:build unix
// +build unix

/*
   Copyright The KWasm Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package containerd

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"syscall"

	"github.com/mitchellh/go-ps"
)

var psProcesses = ps.Processes

type defaultRestarter struct{}

func NewDefaultRestarter() Restarter {
	return defaultRestarter{}
}

func (c defaultRestarter) Restart() error {
	// If listing systemd units succeeds, prefer systemctl restart; otherwise kill pid
	if _, err := ListSystemdUnits(); err == nil {
		out, err := nsenterCmd("systemctl", "restart", "containerd").CombinedOutput()
		slog.Debug(string(out))
		if err != nil {
			return fmt.Errorf("unable to restart containerd: %w", err)
		}
	} else {
		pid, err := getPid("containerd")
		if err != nil {
			return err
		}
		slog.Debug("found containerd process", "pid", pid)

		err = syscall.Kill(pid, syscall.SIGHUP)
		if err != nil {
			return fmt.Errorf("failed to send SIGHUP to containerd: %w", err)
		}
	}

	return nil
}

type K0sRestarter struct{}

func (c K0sRestarter) Restart() error {
	// First, collect systemd units to determine which mode k0s is running in, eg
	// k0sworker or k0scontroller
	units, err := ListSystemdUnits()
	if err != nil {
		return fmt.Errorf("unable to list systemd units: %w", err)
	}
	service := regexp.MustCompile("k0sworker|k0scontroller").FindString(string(units))

	out, err := nsenterCmd("systemctl", "restart", service).CombinedOutput()
	slog.Debug(string(out))
	if err != nil {
		return fmt.Errorf("unable to restart %s: %w", service, err)
	}

	return nil
}

type K3sRestarter struct{}

func (c K3sRestarter) Restart() error {
	// This restarter will be used both for stock K3s distros, which use systemd as well as K3d, which does not.

	// If listing systemd units succeeds, prefer systemctl restart; otherwise kill pid
	// First, collect systemd units to determine which k3s service to restart
	// TODO: It appears the service name itself can be customized, so we may want to consider similar support
	// See https://github.com/k3s-io/k3s/blob/main/install.sh
	if units, err := ListSystemdUnits(); err == nil {
		var service string
		// Prioritize k3s-agent (more common); otherwise k3s
		switch {
		case bytes.Contains(units, []byte("k3s-agent.service")):
			service = "k3s-agent"
		case bytes.Contains(units, []byte("k3s.service")):
			service = "k3s"
		default:
			return fmt.Errorf("failed to find a registered k3s systemd service")
		}

		out, err := nsenterCmd("systemctl", "restart", service).CombinedOutput()
		slog.Debug(string(out))
		if err != nil {
			return fmt.Errorf("unable to restart the %s systemd service: %w", service, err)
		}
	} else {
		// TODO: this approach still leads to the behavior mentioned in https://github.com/spinframework/runtime-class-manager/issues/140:
		// The first pod's provisioner container exits with code 255, leading to pod status Unknown,
		// followed by the subsequent pod's provisioner container no-op-ing and finishing with status Completed.
		pid, err := getPid("k3s")
		if err != nil {
			return err
		}
		slog.Debug("found k3s process", "pid", pid)

		err = syscall.Kill(pid, syscall.SIGHUP)
		if err != nil {
			return fmt.Errorf("failed to send SIGHUP to k3s: %w", err)
		}
	}

	return nil
}

type MicroK8sRestarter struct{}

func (c MicroK8sRestarter) Restart() error {
	out, err := nsenterCmd("systemctl", "restart", "snap.microk8s.daemon-containerd").CombinedOutput()
	slog.Debug(string(out))
	if err != nil {
		return fmt.Errorf("unable to restart snap.microk8s.daemon-containerd: %w", err)
	}

	return nil
}

type RKE2Restarter struct{}

func (c RKE2Restarter) Restart() error {
	// First, collect systemd units to determine which mode rke2 is running in, eg
	// rke2-agent or rke2-server
	units, err := ListSystemdUnits()
	if err != nil {
		return fmt.Errorf("unable to list systemd units: %w", err)
	}
	service := regexp.MustCompile("rke2-agent|rke2-server").FindString(string(units))

	out, err := nsenterCmd("systemctl", "restart", service).CombinedOutput()
	slog.Debug(string(out))
	if err != nil {
		return fmt.Errorf("unable to restart %s: %w", service, err)
	}

	return nil
}

func ListSystemdUnits() ([]byte, error) {
	return nsenterCmd("systemctl", "list-units", "--type", "service").CombinedOutput()
}

func nsenterCmd(cmd ...string) *exec.Cmd {
	// #nosec G204 G702
	return exec.CommandContext(context.Background(), "nsenter",
		append([]string{fmt.Sprintf("-m/%s/proc/1/ns/mnt", os.Getenv("HOST_ROOT")), "--"}, cmd...)...)
}

func getPid(executable string) (int, error) {
	processes, err := psProcesses()
	if err != nil {
		return 0, fmt.Errorf("could not get processes: %w", err)
	}

	var containerdProcesses = []ps.Process{}

	for _, process := range processes {
		if process.Executable() == executable {
			containerdProcesses = append(containerdProcesses, process)
		}
	}

	if len(containerdProcesses) != 1 {
		return 0, fmt.Errorf("need exactly one %s process, found: %d", executable, len(containerdProcesses))
	}

	return containerdProcesses[0].Pid(), nil
}
