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
	"os"
	"path"
	"strings"

	"github.com/spf13/afero"
	"github.com/spinframework/runtime-class-manager/internal/shim"
)

type Restarter interface {
	Restart() error
}

type Config struct {
	hostFs         afero.Fs
	configPath     string
	restarter      Restarter
	runtimeOptions map[string]string
}

func NewConfig(hostFs afero.Fs, configPath string, restarter Restarter, runtimeOptions map[string]string) *Config {
	return &Config{
		hostFs:         hostFs,
		configPath:     configPath,
		restarter:      restarter,
		runtimeOptions: runtimeOptions,
	}
}

func (c *Config) AddRuntime(shimPath string) error {
	runtimeName := shim.RuntimeName(path.Base(shimPath))
	l := slog.With("runtime", runtimeName)

	// Containerd config file needs to exist, otherwise return the error
	data, err := afero.ReadFile(c.hostFs, c.configPath)
	if err != nil {
		return err
	}

	// Warn if config.toml already contains runtimeName
	if strings.Contains(string(data), runtimeName) {
		l.Info("runtime config already exists, skipping")
		return nil
	}

	cfg := generateConfig(shimPath, runtimeName, c.runtimeOptions, data)

	// Open file in append mode
	file, err := c.hostFs.OpenFile(c.configPath, os.O_APPEND|os.O_WRONLY, 0o644) //nolint:mnd // file permissions
	if err != nil {
		return err
	}
	defer file.Close()

	// Append config
	_, err = file.WriteString(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) RemoveRuntime(shimPath string) (changed bool, err error) {
	runtimeName := shim.RuntimeName(path.Base(shimPath))
	l := slog.With("runtime", runtimeName)

	// Containerd config file needs to exist, otherwise return the error
	data, err := afero.ReadFile(c.hostFs, c.configPath)
	if err != nil {
		return false, err
	}

	// Warn if config.toml does not contain the runtimeName
	if !strings.Contains(string(data), runtimeName) {
		l.Warn("runtime config does not exist, skipping")
		return false, nil
	}

	cfg := generateConfig(shimPath, runtimeName, c.runtimeOptions, data)

	// Convert the file data to a string and replace the target string with an empty string.
	modifiedData := strings.ReplaceAll(string(data), cfg, "")

	// Write the modified data back to the file.
	err = afero.WriteFile(c.hostFs, c.configPath, []byte(modifiedData), 0o644) //nolint:mnd // file permissions
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Config) RestartRuntime() error {
	return c.restarter.Restart()
}

func generateConfig(shimPath string, runtimeName string, runtimeOptions map[string]string, configData []byte) string {
	// Config domain for containerd 1.0 (config version 2)
	domain := "io.containerd.grpc.v1.cri"
	if strings.Contains(string(configData), "version = 3") {
		// Config domain for containerd 2.0 (config version 3)
		domain = "io.containerd.cri.v1.runtime"
	}

	runtimeConfiguration := fmt.Sprintf(`
# RCM runtime config for %s
[plugins."%s".containerd.runtimes.%s]
runtime_type = "%s"
`, runtimeName, domain, runtimeName, shimPath)
	// Add runtime options if any are provided
	if len(runtimeOptions) > 0 {
		options := fmt.Sprintf(`[plugins."%s".containerd.runtimes.%s.options]`, domain, runtimeName)
		for k, v := range runtimeOptions {
			options += fmt.Sprintf(`
%s = %s`, k, v)
		}
		runtimeConfiguration += options
	}
	return runtimeConfiguration
}
