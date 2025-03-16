package shim

import (
	"errors"
	"fmt"
	"os"

	"github.com/spinframework/runtime-class-manager/internal/state"
)

func (c *Config) Uninstall(shimName string) (string, error) {
	st, err := state.Get(c.hostFs, c.rcmPath)
	if err != nil {
		return "", err
	}
	s, ok := st.Shims[shimName]
	if !ok {
		return "", fmt.Errorf("shim %s not installed", shimName)
	}
	filePath := s.Path

	err = c.hostFs.Remove(filePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("shim binary at %s does not exist, nothing to delete", filePath)
		}
	}
	st.RemoveShim(shimName)
	if err = st.Write(); err != nil {
		return "", err
	}
	return filePath, err
}
