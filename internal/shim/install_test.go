package shim //nolint:testpackage // whitebox test

import (
	"testing"

	"github.com/spf13/afero"
	tests "github.com/spinframework/runtime-class-manager/tests/node-installer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Install(t *testing.T) {
	type wants struct {
		filepath string
		changed  bool
	}
	type fields struct {
		rootFs    afero.Fs
		hostFs    afero.Fs
		assetPath string
		rcmPath   string
	}
	type args struct {
		shimName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    wants
		wantErr bool
	}{
		{
			"successful shim installation",
			fields{
				tests.FixtureFs("../../testdata/node-installer"),
				afero.NewMemMapFs(),
				"/assets",
				"/opt/rcm",
			},
			args{"containerd-shim-slight-v1"},
			wants{
				"/opt/rcm/bin/containerd-shim-slight-v1",
				true,
			},
			false,
		},
		{
			"no changes to shim",
			fields{
				tests.FixtureFs("../../testdata/node-installer"),
				tests.FixtureFs("../../testdata/node-installer/shim"),
				"/assets",
				"/opt/rcm",
			},
			args{"containerd-shim-spin-v1"},
			wants{
				"/opt/rcm/bin/containerd-shim-spin-v1",
				false,
			},
			false,
		},
		{
			"install new shim over old",
			fields{
				tests.FixtureFs("../../testdata/node-installer"),
				tests.FixtureFs("../../testdata/node-installer/shim"),
				"/assets",
				"/opt/rcm",
			},
			args{"containerd-shim-slight-v1"},
			wants{
				"/opt/rcm/bin/containerd-shim-slight-v1",
				true,
			},
			false,
		},
		{
			"unable to find new shim",
			fields{
				afero.NewMemMapFs(),
				tests.FixtureFs("../../testdata/node-installer/shim"),
				"/assets",
				"/opt/rcm",
			},
			args{"some-shim"},
			wants{
				"",
				false,
			},
			true,
		},
		{
			"unable to write to hostFs",
			fields{
				tests.FixtureFs("../../testdata/node-installer"),
				afero.NewReadOnlyFs(tests.FixtureFs("../../testdata/node-installer/shim")),
				"/assets",
				"/opt/rcm",
			},
			args{"containerd-shim-spin-v1"},
			wants{
				"/opt/rcm/bin/containerd-shim-spin-v1",
				false,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				rootFs:    tt.fields.rootFs,
				hostFs:    tt.fields.hostFs,
				assetPath: tt.fields.assetPath,
				rcmPath:   tt.fields.rcmPath,
			}

			filepath, changed, err := c.Install(tt.args.shimName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.filepath, filepath)
			assert.Equal(t, tt.want.changed, changed)
		})
	}
}
