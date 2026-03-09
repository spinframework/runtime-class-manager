package shim

import (
	"strings"

	"github.com/spf13/afero"
)

type Config struct {
	rootFs    afero.Fs
	hostFs    afero.Fs
	assetPath string
	rcmPath   string
}

func NewConfig(rootFs afero.Fs, hostFs afero.Fs, assetPath string, rcmPath string) *Config {
	return &Config{
		rootFs:    rootFs,
		hostFs:    hostFs,
		assetPath: assetPath,
		rcmPath:   rcmPath,
	}
}

func RuntimeName(bin string) string {
	return strings.TrimPrefix(bin, "containerd-shim-")
}
