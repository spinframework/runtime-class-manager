package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

type State struct {
	Shims        map[string]*Shim `json:"shims"`
	fs           afero.Fs
	lockFilePath string
}

func Get(fs afero.Fs, rcmPath string) (*State, error) {
	out := State{
		Shims:        make(map[string]*Shim),
		lockFilePath: filepath.Join(rcmPath, "rcm-lock.json"),
		fs:           fs,
	}
	content, err := afero.ReadFile(fs, out.lockFilePath)
	if err == nil {
		err := json.Unmarshal(content, &out)
		return &out, err
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &out, nil
}

func (l *State) ShimChanged(shimName string, sha256 []byte, path string) bool {
	shim, ok := l.Shims[shimName]
	if !ok {
		return true
	}

	return !bytes.Equal(shim.Sha256, sha256) || shim.Path != path
}

func (l *State) UpdateShim(shimName string, shim Shim) {
	l.Shims[shimName] = &shim
}

func (l *State) RemoveShim(shimName string) {
	delete(l.Shims, shimName)
}

func (l *State) Write() error {
	out, err := json.MarshalIndent(l, "", " ")
	if err != nil {
		return err
	}

	slog.Debug("writing lock file", "content", string(out))

	return afero.WriteFile(l.fs, l.lockFilePath, out, 0644) //nolint:mnd // file permissions
}
