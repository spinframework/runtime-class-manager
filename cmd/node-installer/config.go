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
