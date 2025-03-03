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

package main_test

import (
	"reflect"
	"testing"

	"github.com/spf13/afero"
	main "github.com/spinkube/runtime-class-manager/cmd/node-installer"
	"github.com/spinkube/runtime-class-manager/internal/preset"
	tests "github.com/spinkube/runtime-class-manager/tests/node-installer"
	"github.com/stretchr/testify/require"
)

func Test_DetectDistro(t *testing.T) {
	type args struct {
		config main.Config
		hostFs afero.Fs
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantPreset preset.Settings
	}{
		{
			"config_override",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", preset.MicroK8s.ConfigPath},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/default"),
			},
			false,
			preset.MicroK8s,
		},
		{
			"config_not_found_fallback_default",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", "/etc/containerd/not_found.toml"},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/default"),
			},
			false,
			preset.Default.WithConfigPath("/etc/containerd/not_found.toml"),
		},
		{
			"default_and_distro_specific_exist",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", ""},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/containerd/default-and-k0s-configs"),
			},
			false,
			preset.K0s,
		},
		{
			"unsupported",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", ""},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/unsupported"),
			},
			true,
			preset.Default,
		},
		{
			"microk8s",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", ""},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/microk8s"),
			},
			false,
			preset.MicroK8s,
		},
		{
			"k0s",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", ""},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/k0s"),
			},
			false,
			preset.K0s,
		},
		{
			"k3s",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", ""},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/k3s"),
			},
			false,
			preset.K3s,
		},
		{
			"rke2",
			args{
				main.Config{
					struct {
						Name       string
						ConfigPath string
					}{"containerd", ""},
					struct {
						Path      string
						AssetPath string
					}{"/opt/kwasm", "/assets"},
					struct{ RootPath string }{""},
				},
				tests.FixtureFs("../../testdata/node-installer/distros/rke2"),
			},
			false,
			preset.RKE2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset, err := main.DetectDistro(tt.args.config, tt.args.hostFs)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantPreset.ConfigPath, preset.ConfigPath)
				require.Equal(t, reflect.ValueOf(tt.wantPreset.Setup), reflect.ValueOf(preset.Setup))
				require.Equal(t, reflect.ValueOf(tt.wantPreset.Restarter), reflect.ValueOf(preset.Restarter))
			}
		})
	}
}
