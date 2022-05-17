package pkg

import (
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