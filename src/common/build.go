package common

import (
	"gopkg.in/yaml.v2"

	"io/ioutil"
	"path/filepath"
)

type Build struct {
	WorkingDir    string `db:"-"`
	BuildID       int64  `db:"build_id"`
	Branch        string `db:"branch"`
	Revision      string `db:"revision"`
	RepositoryURL string `db:"repository_url"`
	Config        struct {
		Images   []string `yaml:"images"`
		Commands []string `yaml:"commands"`
		Postgres struct {
			Database string `yaml:"database"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `postgres`
	}
}

func (b *Build) LoadConfig() error {

	data, err := ioutil.ReadFile(filepath.Join(b.WorkingDir, ".postgres-ci.yaml"))

	if err != nil {

		return nil
	}

	if err := yaml.Unmarshal(data, &b.Config); err != nil {

		return err
	}

	return nil
}
