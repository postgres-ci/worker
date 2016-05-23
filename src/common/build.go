package common

import (
	"gopkg.in/yaml.v2"

	"fmt"
	"io/ioutil"
	"path/filepath"
)

type Build struct {
	WorkingDir    string `db:"-"`
	BuildID       int32  `db:"build_id"`
	Branch        string `db:"branch"`
	Revision      string `db:"revision"`
	RepositoryURL string `db:"repository_url"`
	RawConfig     string
	Config        struct {
		Env        []string `yaml:"env"`
		Tests      []string `yaml:"tests"`
		Images     []string `yaml:"images"`
		Commands   []string `yaml:"commands"`
		Entrypoint []string `yaml:"entrypoint"`
		Postgres   struct {
			Database string `yaml:"database"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `postgres`
	}
}

func (b *Build) DSN(host string) string {

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host,
		b.Config.Postgres.Username,
		b.Config.Postgres.Password,
		b.Config.Postgres.Database,
	)
}

func (b *Build) LoadConfig() error {

	data, err := ioutil.ReadFile(filepath.Join(b.WorkingDir, ".postgres-ci.yaml"))

	if err != nil {

		return fmt.Errorf("Could not read .postgres-ci.yaml file. %v", err)
	}

	if err := yaml.Unmarshal(data, &b.Config); err != nil {

		return err
	}

	b.RawConfig = string(data)

	return nil
}
