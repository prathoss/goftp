package pkg

import (
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/prathoss/goftp/types"
)

func OsDeleteFn(location string, entries []types.Entry) error {
	for _, entry := range entries {
		absolutePath := path.Join(location, entry.Name)
		var delFn func(string) error
		switch entry.Type {
		case types.TypeDirectory:
			delFn = os.RemoveAll
		default:
			delFn = os.Remove
		}
		if err := delFn(absolutePath); err != nil {
			return err
		}
	}
	return nil
}

func createDirIfNotExist(path string) error {
	if err := os.Mkdir(path, 0750); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return nil
}
