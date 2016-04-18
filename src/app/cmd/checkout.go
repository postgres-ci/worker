package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/git"
)

type Checkout struct {
	WorkingDir string
}

func (c *Checkout) Run(build *common.Build) error {

	path, err := git.CheckoutToRevision(build.RepositoryURL, c.WorkingDir, build.Branch, build.Revision)

	if err != nil {

		log.Errorf("Could not checkout repo: %v", err)

		return err
	}

	log.Debugf("Checkout '%s' into '%s'", build.RepositoryURL, path)
	log.Debugf("Branch: %s, revision: %s", build.Branch, build.Revision)

	build.WorkingDir = path

	if err := build.LoadConfig(); err != nil {

		log.Errorf("Could not load build config: %v", err)

		return err
	}

	return nil
}
