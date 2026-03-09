package shim

import (
	"crypto/sha256"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/spinframework/runtime-class-manager/internal/state"
)

func (c *Config) Install(shimName string) (filePath string, changed bool, err error) {
	shimPath := filepath.Join(c.assetPath, shimName)
	srcFile, err := c.rootFs.OpenFile(shimPath, os.O_RDONLY, 0o000) //nolint:mnd // file permissions
	if err != nil {
		return "", false, err
	}
	dstFilePath := path.Join(c.rcmPath, "bin", shimName)

	err = c.hostFs.MkdirAll(path.Dir(dstFilePath), 0o775) //nolint:mnd // file permissions
	if err != nil {
		return dstFilePath, false, err
	}

	dstFile, err := c.hostFs.OpenFile(dstFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755) //nolint:mnd // file permissions
	if err != nil {
		return "", false, err
	}

	st, err := state.Get(c.hostFs, c.rcmPath)
	if err != nil {
		return "", false, err
	}
	shimSha256 := sha256.New()

	_, err = io.Copy(io.MultiWriter(dstFile, shimSha256), srcFile)
	runtimeName := RuntimeName(shimName)
	changed = st.ShimChanged(runtimeName, shimSha256.Sum(nil), dstFilePath)
	if changed {
		st.UpdateShim(runtimeName, state.Shim{
			Path:   dstFilePath,
			Sha256: shimSha256.Sum(nil),
		})
		if err := st.Write(); err != nil {
			return "", false, err
		}
	}

	return dstFilePath, changed, err
}
