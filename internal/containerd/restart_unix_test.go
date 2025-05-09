//go:build unix
// +build unix

package containerd //nolint:testpackage // whitebox test

import (
	"fmt"
	"testing"

	"github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProcess struct {
	executable string
	pid        int
}

func (p *mockProcess) Executable() string {
	return p.executable
}

func (p *mockProcess) Pid() int {
	return p.pid
}

func (p *mockProcess) PPid() int {
	return 0
}

func Test_getPid(t *testing.T) {
	tests := []struct {
		name             string
		psProccessesMock func() ([]ps.Process, error)
		want             int
		wantErr          bool
	}{
		{"no containerd process found", func() ([]ps.Process, error) {
			return []ps.Process{}, nil
		}, 0, true},
		{"single containerd process found", func() ([]ps.Process, error) {
			return []ps.Process{
				&mockProcess{executable: "containerd", pid: 123},
			}, nil
		}, 123, false},
		{"multiple containerd processes found", func() ([]ps.Process, error) {
			return []ps.Process{
				&mockProcess{executable: "containerd", pid: 0},
				&mockProcess{executable: "containerd", pid: 0},
			}, nil
		}, 0, true},
		{"error getting processes", func() ([]ps.Process, error) {
			return nil, fmt.Errorf("error getting processes")
		}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psProcesses = tt.psProccessesMock
			got, err := getPid("containerd")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
