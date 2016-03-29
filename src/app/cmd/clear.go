package cmd

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/common"

	"os"
)

type Clear struct{}

func (c *Clear) Run(build *common.Build) error {

	logger.Debugf("Remove dir '%s'", build.WorkingDir)

	if err := os.RemoveAll(build.WorkingDir); err != nil {

		logger.Errorf("Could not remove dir '%s'. Err %v", build.WorkingDir, err)
	}

	return nil
}
