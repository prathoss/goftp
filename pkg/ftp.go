package pkg

import (
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/jlaffaye/ftp"
	"github.com/prathoss/goftp/types"
)

func PrepareDownloadFn(client *ftp.ServerConn) func(string, []types.Entry, string) error {
	return func(root string, entries []types.Entry, destination string) error {
		for _, entry := range entries {
			absoluteSourcePath := path.Join(root, entry.Name)
			err := func() error {
				if entry.Type == types.TypeDirectory {
					walker := client.Walk(absoluteSourcePath)
					for walker.Next() {
						if walker.Stat().Type != ftp.EntryTypeFile {
							continue
						}
						// TODO: finish implementation
						//  * check if directory is created
						// downloadFile(client)
					}
					return nil
				}
				if err := downloadFile(client, absoluteSourcePath, path.Join(destination, entry.Name)); err != nil {
					return err
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func downloadFile(c *ftp.ServerConn, source, destination string) error {
	result, err := c.Retr(source)
	if err != nil {
		return err
	}
	defer result.Close()
	fl, err := os.OpenFile(destination, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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
			absoluteSourcePath := path.Join(root, entry.Name)
			err := func() error {
				if entry.Type == types.TypeDirectory {
					return filepath.Walk(
						absoluteSourcePath,
						func(path string, info fs.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if info.IsDir() {
								return nil
							}
							// TODO: finish implementation
							//  * check if directory is created
							// uploadFile(client)
							return nil
						})
				}
				if err := uploadFile(client, absoluteSourcePath, path.Join(destination, entry.Name)); err != nil {
					return err
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}
		return nil
	}
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
