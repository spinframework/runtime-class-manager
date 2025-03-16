package preset_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spinframework/runtime-class-manager/internal/preset"
	tests "github.com/spinframework/runtime-class-manager/tests/node-installer"
	"github.com/stretchr/testify/require"
)

func Test_WithSetup(t *testing.T) {
	type args struct {
		settings preset.Settings
		hostFs   afero.Fs
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantContents string
	}{
		{
			"rke2_err",
			args{
				preset.RKE2,
				tests.FixtureFs("../../testdata/node-installer/distros/unsupported"),
			},
			true,
			"",
		},
		{
			"rke2_config_exists",
			args{
				preset.RKE2,
				tests.FixtureFs("../../testdata/node-installer/containerd/rke2-existing-config-tmpl"),
			},
			false,
			"version = 2\npreexisting-config = true",
		},
		{
			"rke2_config_is_created",
			args{
				preset.RKE2,
				tests.FixtureFs("../../testdata/node-installer/distros/rke2"),
			},
			false,
			"version = 2",
		},
		{
			"k3s",
			args{
				preset.K3s,
				tests.FixtureFs("../../testdata/node-installer/distros/k3s"),
			},
			false,
			"version = 2",
		},
		{
			"k0s",
			args{
				preset.K0s,
				tests.FixtureFs("../../testdata/node-installer/distros/k0s"),
			},
			false,
			"",
		},
		{
			"microk8s",
			args{
				preset.MicroK8s,
				tests.FixtureFs("../../testdata/node-installer/distros/microk8s"),
			},
			false,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.settings.Setup(
				preset.Env{
					ConfigPath: tt.args.settings.ConfigPath,
					HostFs:     tt.args.hostFs,
				},
			)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				bytes, err := afero.ReadFile(tt.args.hostFs, tt.args.settings.ConfigPath)
				require.NoError(t, err)
				require.Equal(t, tt.wantContents, string(bytes))
			}
		})
	}
}
