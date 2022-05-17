package pkg

import (
	"errors"
	"io"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type ServerConf struct {
	Server string
	Port   int
	User   string
}

type Conf struct {
	Servers []ServerConf
}

func (c Conf) ServerExists(newServer ServerConf) bool {
	for _, server := range c.Servers {
		if server == newServer {
			return true
		}
	}
	return false
}

func getConfFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".goftp.yaml"), nil
}

func getConfFile() (*os.File, error) {
	filePath, err := getConfFilePath()
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func readConfig(r io.Reader) (Conf, error) {
	var cfg Conf
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return Conf{}, err
	}
	return cfg, nil
}

func writeConfig(cfg Conf, w io.WriteSeeker) error {
	if _, err := w.Seek(0, 0); err != nil {
		return err
	}
	if err := yaml.NewEncoder(w).Encode(cfg); err != nil {
		return err
	}
	return nil
}

func GetConfig() (Conf, error) {
	cf, err := getConfFile()
	if err != nil {
		return Conf{}, err
	}
	defer cf.Close()
	c, err := readConfig(cf)
	if err != nil {
		return Conf{}, err
	}
	return c, nil
}

func AddToConfig(conf ServerConf) error {
	f, err := getConfFile()
	if err != nil {
		return err
	}
	defer f.Close()
	c, err := readConfig(f)
	if err != nil {
		return err
	}
	if c.ServerExists(conf) {
		return nil
	}
	c.Servers = append(c.Servers, conf)
	if err := writeConfig(c, f); err != nil {
		return err
	}
	return nil
}
