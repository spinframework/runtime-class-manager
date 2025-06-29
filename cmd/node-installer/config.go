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

package main

type Config struct {
	Runtime struct {
		Name       string
		ConfigPath string
		// Options is a map of containerd runtime options for the shim plugin.
		// See an example of the cgroup drive option here:
		// https://github.com/containerd/containerd/blob/main/docs/cri/config.md#cgroup-driver
		Options map[string]string
	}
	RCM struct {
		Path      string
		AssetPath string
	}
	Host struct {
		RootPath string
	}
}
