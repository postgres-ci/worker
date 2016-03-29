package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/common"
)

const (
	RUN_BUILD_SQL = `
		SELECT 
			build_id,
			branch,
			revision,
			repository_url 
		FROM postgres_ci.run_build($1)
	`
)

func (a *app) execute() {

	for {

		task := <-a.tasks

		log.Debugf("Task build_id [%d] at %s", task.BuildID, task.CreatedAt)

		var build common.Build

		if err := a.connect.Get(&build, RUN_BUILD_SQL, task.BuildID); err != nil {

			log.Errorf("Could not run build (%v)", err)

			continue
		}

		if err := build.LoadConfig(); err != nil {

			log.Errorf("Could not load build config (%v)", err)

			continue
		}

		for _, cmd := range a.commands {

			if err := cmd.Run(&build); err != nil {

				break
			}
		}

		log.Debug(build)
	}
}
