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

package containerd //nolint:testpackage // whitebox test

import (
	"testing"

	"github.com/spf13/afero"
	tests "github.com/spinframework/runtime-class-manager/tests/node-installer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_AddRuntime(t *testing.T) {
	type fields struct {
		hostFs     afero.Fs
		configPath string
	}
	type args struct {
		shimPath string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantFileContent string
	}{
		{"missing shim config", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/missing-containerd-shim-config"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, false, `[plugins]
  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false
  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]
  [plugins."io.containerd.gc.v1.scheduler"]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = 0
    startup_delay = "100ms"
  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = true
  [plugins."io.containerd.service.v1.tasks-service"]
    blockio_config_file = ""
    rdt_config_file = ""

# RCM runtime config for spin-v1
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.spin-v1]
runtime_type = "/opt/rcm/bin/containerd-shim-spin-v1"
`},
		{"missing config", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/missing-containerd-config"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, true, ``},
		{"existing shim config", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/existing-containerd-shim-config"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, false, `[plugins]
  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false
  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]
  [plugins."io.containerd.gc.v1.scheduler"]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = 0
    startup_delay = "100ms"
  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = true
  [plugins."io.containerd.service.v1.tasks-service"]
    blockio_config_file = ""
    rdt_config_file = ""

# RCM runtime config for spin-v1
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.spin-v1]
runtime_type = "/opt/rcm/bin/containerd-shim-spin-v1"
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				hostFs:     tt.fields.hostFs,
				configPath: tt.fields.configPath,
			}
			err := c.AddRuntime(tt.args.shimPath)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			gotContent, err := afero.ReadFile(c.hostFs, c.configPath)
			require.NoError(t, err)

			assert.Equal(t, tt.wantFileContent, string(gotContent))
		})
	}
}

func TestConfig_AddRuntimeOptions(t *testing.T) {
	wantFileContent := `[plugins]
  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false
  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]
  [plugins."io.containerd.gc.v1.scheduler"]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = 0
    startup_delay = "100ms"
  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = true
  [plugins."io.containerd.service.v1.tasks-service"]
    blockio_config_file = ""
    rdt_config_file = ""

# RCM runtime config for spin-v1
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.spin-v1]
runtime_type = "/opt/rcm/bin/containerd-shim-spin-v1"
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.spin-v1.options]
SystemdCgroup = true`
	t.Run("plugin options added", func(t *testing.T) {
		c := &Config{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/missing-containerd-shim-config"),
			configPath: "/etc/containerd/config.toml",
			runtimeOptions: map[string]string{
				"SystemdCgroup": "true",
			},
		}
		err := c.AddRuntime("/opt/rcm/bin/containerd-shim-spin-v1")

		require.NoError(t, err)

		gotContent, err := afero.ReadFile(c.hostFs, c.configPath)
		require.NoError(t, err)

		assert.Equal(t, wantFileContent, string(gotContent))
	})
}

func TestConfig_RemoveRuntime(t *testing.T) {
	type fields struct {
		hostFs     afero.Fs
		configPath string
	}
	type args struct {
		shimPath string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantFileContent string
	}{
		{"missing shim config", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/missing-containerd-shim-config"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, false, `[plugins]
  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false
  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]
  [plugins."io.containerd.gc.v1.scheduler"]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = 0
    startup_delay = "100ms"
  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = true
  [plugins."io.containerd.service.v1.tasks-service"]
    blockio_config_file = ""
    rdt_config_file = ""
`},
		{"missing config", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/missing-containerd-config"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, true, ``},
		{"existing shim config", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/existing-containerd-shim-config"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, false, `[plugins]
  [plugins."io.containerd.monitor.v1.cgroups"]
    no_prometheus = false
  [plugins."io.containerd.service.v1.diff-service"]
    default = ["walking"]
  [plugins."io.containerd.gc.v1.scheduler"]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = 0
    startup_delay = "100ms"
  [plugins."io.containerd.runtime.v2.task"]
    platforms = ["linux/amd64"]
    sched_core = true
  [plugins."io.containerd.service.v1.tasks-service"]
    blockio_config_file = ""
    rdt_config_file = ""
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				hostFs:     tt.fields.hostFs,
				configPath: tt.fields.configPath,
			}
			_, err := c.RemoveRuntime(tt.args.shimPath)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			gotContent, err := afero.ReadFile(c.hostFs, c.configPath)
			require.NoError(t, err)

			assert.Equal(t, tt.wantFileContent, string(gotContent))
		})
	}
}

func TestGenerateConfig_ContainerdVersions(t *testing.T) {
	type fields struct {
		hostFs     afero.Fs
		configPath string
	}
	type args struct {
		shimPath string
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantFileContent string
	}{
		{"containerd 1.x", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/1.x"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, false, `version = 2
# RCM runtime config for spin-v1
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.spin-v1]
runtime_type = "/opt/rcm/bin/containerd-shim-spin-v1"
`},
		{"containerd 2.x", fields{
			hostFs:     tests.FixtureFs("../../testdata/node-installer/containerd/2.x"),
			configPath: "/etc/containerd/config.toml",
		}, args{"/opt/rcm/bin/containerd-shim-spin-v1"}, false, `version = 3
# RCM runtime config for spin-v1
[plugins."io.containerd.cri.v1.runtime".containerd.runtimes.spin-v1]
runtime_type = "/opt/rcm/bin/containerd-shim-spin-v1"
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				hostFs:     tt.fields.hostFs,
				configPath: tt.fields.configPath,
			}
			err := c.AddRuntime(tt.args.shimPath)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			gotContent, err := afero.ReadFile(c.hostFs, c.configPath)
			require.NoError(t, err)

			assert.Equal(t, tt.wantFileContent, string(gotContent))
		})
	}
}
