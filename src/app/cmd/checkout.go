package cmd

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/git"
)

type Checkout struct {
	WorkingDir string
}

func (c *Checkout) Run(build *common.Build) error {

	path, err := git.CheckoutToRevision(build.RepositoryURL, c.WorkingDir, build.Branch, build.Revision)

	if err != nil {

		(&Clear{}).Run(build)

		return err
	}

	logger.Debugf("Checkout %s into %s", build.RepositoryURL, path)
	logger.Debugf("Branch: %s, revision: %s", build.Branch, build.Revision)

	build.WorkingDir = path

	return nil
}
