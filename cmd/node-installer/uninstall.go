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

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/spinframework/runtime-class-manager/internal/containerd"
	"github.com/spinframework/runtime-class-manager/internal/shim"
)

// uninstallCmd represents the uninstall command.
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall containerd shims",
	Run: func(_ *cobra.Command, _ []string) {
		rootFs := afero.NewOsFs()
		hostFs := afero.NewBasePathFs(rootFs, config.Host.RootPath)

		distro, err := DetectDistro(config, hostFs)
		if err != nil {
			slog.Error("failed to detect containerd config", "error", err)
			os.Exit(1)
		}

		config.Runtime.ConfigPath = distro.ConfigPath

		if err := RunUninstall(config, rootFs, hostFs, distro.Restarter); err != nil {
			slog.Error("failed to uninstall", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func RunUninstall(config Config, rootFs, hostFs afero.Fs, restarter containerd.Restarter) error {
	slog.Info("uninstall called", "shim", config.Runtime.Name)
	shimName := config.Runtime.Name
	runtimeName := path.Join(config.RCM.Path, "bin", shimName)

	containerdConfig := containerd.NewConfig(hostFs, config.Runtime.ConfigPath, restarter)
	shimConfig := shim.NewConfig(rootFs, hostFs, config.RCM.AssetPath, config.RCM.Path)

	binPath, err := shimConfig.Uninstall(shimName)
	if err != nil {
		return fmt.Errorf("failed to delete shim '%s': %w", runtimeName, err)
	}

	configChanged, err := containerdConfig.RemoveRuntime(binPath)
	if err != nil {
		return fmt.Errorf("failed to write containerd config for shim '%s': %w", runtimeName, err)
	}

	if !configChanged {
		slog.Info("nothing changed, nothing more to do")
		return nil
	}

	slog.Info("restarting containerd")
	err = containerdConfig.RestartRuntime()
	if err != nil {
		return fmt.Errorf("failed to restart containerd: %w", err)
	}

	return nil
}
