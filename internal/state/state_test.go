package state_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spinframework/runtime-class-manager/internal/state"
	tests "github.com/spinframework/runtime-class-manager/tests/node-installer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	type args struct {
		fs      afero.Fs
		rcmPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *state.State
		wantErr bool
	}{
		{
			"existing state",
			args{
				tests.FixtureFs("../../testdata/node-installer/containerd/existing-containerd-shim-config"),
				"/opt/rcm",
			},
			&state.State{
				Shims: map[string]*state.Shim{
					"spin-v1": {
						Sha256: []byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
						Path:   "/opt/rcm/bin/containerd-shim-spin-v1",
					},
				},
			},
			false,
		},
		{
			"missing state",
			args{
				tests.FixtureFs("../../testdata/node-installer/containerd/missing-containerd-shim-config"),
				"/opt/rcm",
			},
			&state.State{
				Shims: map[string]*state.Shim{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := state.Get(tt.args.fs, tt.args.rcmPath)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want.Shims, got.Shims)
		})
	}
}

func TestShimChanged(t *testing.T) {
	type args struct {
		shimName string
		sha256   []byte
		path     string
	}
	state := &state.State{
		Shims: map[string]*state.Shim{
			"spin-v1": {
				Sha256: []byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
				Path:   "/opt/rcm/bin/containerd-shim-spin-v1",
			},
		},
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"existing shim, same sha256 and path",
			args{
				"spin-v1",
				[]byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
				"/opt/rcm/bin/containerd-shim-spin-v1",
			},
			false,
		},
		{
			"existing shim, different sha256",
			args{
				"spin-v1",
				[]byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 83},
				"/opt/rcm/bin/containerd-shim-spin-v1",
			},
			true,
		},
		{
			"existing shim, different path",
			args{
				"spin-v1",
				[]byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
				"/opt/rcm/bin/containerd-shim-spin-v2",
			},
			true,
		},
		{
			"non-existing shim",
			args{
				"non-existing",
				[]byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
				"/opt/rcm/bin/containerd-shim-spin-v1",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := state.ShimChanged(tt.args.shimName, tt.args.sha256, tt.args.path)
			assert.Equal(t, tt.want, changed)
		})
	}
}
