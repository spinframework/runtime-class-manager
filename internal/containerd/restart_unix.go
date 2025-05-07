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
	"fmt"
	"log/slog"
	"syscall"

	"github.com/mitchellh/go-ps"
)

var psProcesses = ps.Processes

type restarter struct{}

func NewRestarter() Restarter {
	return restarter{}
}

func (c restarter) Restart() error {
	// To make sure any runtime options are applied, ensure D-Bus is
	// installed and running if using systemd
	if UsesSystemd() {
		err := InstallDbus()
		if err != nil {
			return fmt.Errorf("failed to install D-Bus: %w", err)
		}
	}

	pid, err := getPid()
	if err != nil {
		return err
	}
	slog.Debug("found containerd process", "pid", pid)

	err = syscall.Kill(pid, syscall.SIGHUP)

	if err != nil {
		return fmt.Errorf("failed to send SIGHUP to containerd: %w", err)
	}
	return nil
}

func getPid() (int, error) {
	processes, err := psProcesses()
	if err != nil {
		return 0, fmt.Errorf("could not get processes: %w", err)
	}

	var containerdProcesses = []ps.Process{}

	for _, process := range processes {
		if process.Executable() == "containerd" {
			containerdProcesses = append(containerdProcesses, process)
		}
	}

	if len(containerdProcesses) != 1 {
		return 0, fmt.Errorf("need exactly one containerd process, found: %d", len(containerdProcesses))
	}

	return containerdProcesses[0].Pid(), nil
}
