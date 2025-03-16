package tests

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

func FixtureFs(fixturePath string) afero.Fs {
	baseFs := afero.NewBasePathFs(afero.NewOsFs(), fixturePath)
	fs := afero.NewMemMapFs()
	err := afero.Walk(baseFs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return fs.MkdirAll(path, 0755) //nolint:mnd // file permissions
		}
		src, err := baseFs.Open(path)
		if err != nil {
			return err
		}
		dst, err := fs.Create(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return fs
}
