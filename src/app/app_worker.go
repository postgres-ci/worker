package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/app/cmd"
	"github.com/postgres-ci/worker/src/common"
)

const (
	StartBuildSql = `
		SELECT 
			branch,
			revision,
			repository_url 
		FROM build.start($1)
	`
)

func (a *app) worker() {

	for {

		task := <-a.tasks

		var (
			build = common.Build{
				BuildID: task.BuildID,
			}
			buildError string
		)

		if err := a.connect.Get(&build, StartBuildSql, task.BuildID); err != nil {

			log.Errorf("Could not start build: %v", err)

			continue
		}

		for _, command := range a.commands {

			if err := command.Run(&build); err != nil {

				(&cmd.Clear{}).Run(&build)

				buildError = err.Error()

				break
			}
		}

		if _, err := a.connect.Exec(`SELECT build.stop($1, $2, $3)`, build.BuildID, build.RawConfig, buildError); err != nil {

			log.Errorf("Could not stop build: %v", err)
		}
	}
}
