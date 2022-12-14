package pkg

import (
	"errors"
	"io"
	"io/fs"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/prathoss/goftp/types"
)

func PrepareDownloadFn(client *ftp.ServerConn) func(string, []types.Entry, string) error {
	return func(root string, entries []types.Entry, destination string) error {
		for _, entry := range entries {
			if entry.Type == types.TypeDirectory {
				if err := downloadFolderWithContents(client, root, destination, entry); err != nil {
					return err
				}
				continue
			}
			if err := downloadFile(client, path.Join(root, entry.Name), path.Join(destination, entry.Name)); err != nil {
				return err
			}
		}
		return nil
	}
}

func downloadFolderWithContents(client *ftp.ServerConn, root, destination string, entry types.Entry) error {
	if err := createDirIfNotExist(path.Join(destination, entry.Name)); err != nil {
		return err
	}
	walker := client.Walk(path.Join(root, entry.Name))
	for walker.Next() {
		// get relative path => walk prints absolute path when called with absolute
		rel, err := filepath.Rel(root, walker.Path())
		if err != nil {
			return err
		}
		destinationAbs := path.Join(destination, rel)
		if walker.Stat().Type == ftp.EntryTypeFolder {
			if err := createDirIfNotExist(destinationAbs); err != nil {
				return err
			}
			continue
		}
		if err := downloadFile(client, walker.Path(), destinationAbs); err != nil {
			return err
		}
	}
	return nil
}

func downloadFile(c *ftp.ServerConn, source, destination string) error {
	result, err := c.Retr(source)
	if err != nil {
		return err
	}
	defer result.Close()
	fl, err := os.OpenFile(destination, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer fl.Close()
	if _, err := fl.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.Copy(fl, result); err != nil {
		return err
	}
	return nil
}

func PrepareUploadFn(client *ftp.ServerConn) func(string, []types.Entry, string) error {
	return func(root string, entries []types.Entry, destination string) error {
		for _, entry := range entries {
			// func to properly handle defer
			if entry.Type == types.TypeDirectory {
				if err := uploadDirWithContent(client, root, destination, entry); err != nil {
					return err
				}
				continue
			}
			if err := uploadFile(client, path.Join(root, entry.Name), path.Join(destination, entry.Name)); err != nil {
				return err
			}
		}
		return nil
	}
}

func uploadDirWithContent(client *ftp.ServerConn, root, destination string, entry types.Entry) error {
	return filepath.Walk(
		path.Join(root, entry.Name),
		func(walkPath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, walkPath)
			if err != nil {
				return err
			}
			destinationAbs := path.Join(destination, rel)
			if info.IsDir() {
				if err := client.MakeDir(destinationAbs); !isErrorDirExists(err) {
					return err
				}
				return nil
			}
			// walkPath will be absolute => remove root to make destination
			return uploadFile(client, walkPath, destinationAbs)
		})
}

func isErrorDirExists(err error) bool {
	var tpErr *textproto.Error
	return errors.As(err, &tpErr) && tpErr.Code == ftp.StatusFileUnavailable
}

func uploadFile(c *ftp.ServerConn, source, destination string) error {
	fl, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fl.Close()
	if err := c.Stor(destination, fl); err != nil {
		return err
	}
	return nil
}

func PrepareFtpDeleteFn(client *ftp.ServerConn) func(location string, entries []types.Entry) error {
	return func(location string, entries []types.Entry) error {
		for _, entry := range entries {
			absolutePath := path.Join(location, entry.Name)
			var delFn func(string) error
			switch entry.Type {
			case types.TypeDirectory:
				delFn = client.RemoveDirRecur
			default:
				delFn = client.Delete
			}
			if err := delFn(absolutePath); err != nil {
				return err
			}
		}
		return nil
	}
}

func KeepFtpAlive(client *ftp.ServerConn) (chan<- struct{}, <-chan error) {
	ticker := time.NewTimer(15 * time.Second)
	quit := make(chan struct{})
	errChan := make(chan error)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := client.NoOp(); err != nil {
					errChan <- err
					return
				}
			case <-quit:
				return
			}
		}
	}()
	return quit, errChan
}

func FtpToEntry(f *ftp.Entry) types.Entry {
	var tp int
	switch f.Type {
	case ftp.EntryTypeFile:
		tp = types.TypeFile
	case ftp.EntryTypeLink:
		tp = types.TypeLink
	case ftp.EntryTypeFolder:
		tp = types.TypeDirectory
	}
	return types.Entry{
		Name: f.Name,
		Type: tp,
		Size: f.Size,
	}
}
