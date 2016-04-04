package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/common"

	"os"
)

type Clear struct{}

func (c *Clear) Run(build *common.Build) error {

	log.Debugf("Remove dir '%s'", build.WorkingDir)

	if err := os.RemoveAll(build.WorkingDir); err != nil {

		log.Errorf("Could not remove dir '%s': %v", build.WorkingDir, err)
	}

	return nil
}
