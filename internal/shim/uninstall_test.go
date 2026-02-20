package shim //nolint:testpackage // whitebox test

import (
	"testing"

	"github.com/spf13/afero"
	tests "github.com/spinframework/runtime-class-manager/tests/node-installer"
)

func TestConfig_Uninstall(t *testing.T) {
	type fields struct {
		hostFs  afero.Fs
		rcmPath string
	}
	type args struct {
		shimName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			"shim not installed",
			fields{
				tests.FixtureFs("../../testdata/node-installer/shim"),
				"/opt/rcm",
			},
			args{"not-existing-shim"},
			"",
			true,
		},
		{
			"missing shim binary",
			fields{
				tests.FixtureFs("../../testdata/node-installer/shim-missing-binary"),
				"/opt/rcm",
			},
			args{"spin-v1"},
			"/opt/rcm/bin/containerd-shim-spin-v1",
			false,
		},
		{
			"successful shim uninstallation",
			fields{
				tests.FixtureFs("../../testdata/node-installer/shim"),
				"/opt/rcm",
			},
			args{"spin-v1"},
			"/opt/rcm/bin/containerd-shim-spin-v1",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				hostFs:  tt.fields.hostFs,
				rcmPath: tt.fields.rcmPath,
			}

			got, err := c.Uninstall(tt.args.shimName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Uninstall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Config.Uninstall() = %v, want %v", got, tt.want)
			}
		})
	}
}
