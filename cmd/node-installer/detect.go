/*
   Copyright The SpinKube Authors.

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

package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/spf13/afero"
	"github.com/spinframework/runtime-class-manager/internal/preset"
)

var containerdConfigLocations = map[string]preset.Settings{
	// Microk8s
	"/var/snap/microk8s/current/args/containerd-template.toml": preset.MicroK8s,
	// RKE2
	"/var/lib/rancher/rke2/agent/etc/containerd/config.toml": preset.RKE2,
	// K3s
	"/var/lib/rancher/k3s/agent/etc/containerd/config.toml": preset.K3s,
	// K0s
	"/etc/k0s/containerd.toml": preset.K0s,
	// default
	"/etc/containerd/config.toml": preset.Default,
}

func DetectDistro(config Config, hostFs afero.Fs) (preset.Settings, error) {
	if config.Runtime.ConfigPath != "" {
		// containerd config path has been set explicitly
		if distro, ok := containerdConfigLocations[config.Runtime.ConfigPath]; ok {
			return distro, nil
		}
		slog.Warn("could not determine distro from containerd config, falling back to defaults", "config", config.Runtime.ConfigPath)
		return preset.Default.WithConfigPath(config.Runtime.ConfigPath), nil
	}

	var errs []error

	for loc, distro := range containerdConfigLocations {
		_, err := hostFs.Stat(loc)
		if err == nil {
			// config file found, return corresponding distro settings
			return distro, nil
		}
		errs = append(errs, err)
	}

	return preset.Settings{}, fmt.Errorf("failed to detect containerd config path: %w", errors.Join(errs...))
}
