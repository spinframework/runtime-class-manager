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

package shim //nolint:testpackage // whitebox test

import (
	"testing"

	"github.com/spf13/afero"
	tests "github.com/spinkube/runtime-class-manager/tests/node-installer"
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
