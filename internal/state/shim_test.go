package state //nolint:testpackage // whitebox test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShim_MarshalJSON(t *testing.T) {
	type fields struct {
		Sha256 []byte
		Path   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			"default",
			fields{
				Sha256: []byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
				Path:   "/opt/rcm/bin/containerd-shim-spin-v1",
			},
			`{"sha256":"6da5e8f17a9bfa9cb04cf22c87b6475394ecec3af4fdc337f72d6dbf3319ea52","path":"/opt/rcm/bin/containerd-shim-spin-v1"}`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Shim{
				Sha256: tt.fields.Sha256,
				Path:   tt.fields.Path,
			}
			got, err := s.MarshalJSON()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestShim_UnmarshalJSON(t *testing.T) {
	type wants struct {
		path   string
		sha256 []byte
	}
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    wants
		wantErr bool
	}{
		{
			"default",
			args{
				`{"sha256":"6da5e8f17a9bfa9cb04cf22c87b6475394ecec3af4fdc337f72d6dbf3319ea52","path":"/opt/rcm/bin/containerd-shim-spin-v1"}`,
			},
			wants{
				path:   "/opt/rcm/bin/containerd-shim-spin-v1",
				sha256: []byte{109, 165, 232, 241, 122, 155, 250, 156, 176, 76, 242, 44, 135, 182, 71, 83, 148, 236, 236, 58, 244, 253, 195, 55, 247, 45, 109, 191, 51, 25, 234, 82},
			},
			false,
		},
		{
			"broken sha",
			args{
				`{"sha256":"2","path":"/opt/rcm/bin/containerd-shim-spin-v1"}`,
			},
			wants{
				path:   "/opt/rcm/bin/containerd-shim-spin-v1",
				sha256: nil,
			},
			true,
		},
		{
			"broken json",
			args{
				`broken`,
			},
			wants{
				path:   "",
				sha256: nil,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Shim
			err := s.UnmarshalJSON([]byte(tt.args.data))

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want.path, s.Path)
			assert.Equal(t, tt.want.sha256, s.Sha256)
		})
	}
}
