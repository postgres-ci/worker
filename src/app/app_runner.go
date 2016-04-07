package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/app/cmd"
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

func (a *app) runner() {

	for {

		task := <-a.tasks

		var build common.Build

		if err := a.connect.Get(&build, RUN_BUILD_SQL, task.BuildID); err != nil {

			log.Errorf("Could not run build: %v", err)

			continue
		}

		for _, command := range a.commands {

			if err := command.Run(&build); err != nil {

				(&cmd.Clear{}).Run(&build)

				break
			}
		}
	}
}
