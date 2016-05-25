package runner

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"

	"fmt"
	"strings"
	"time"
)

func Custom(image, assets string, dockerClient docker.Client, connect *sqlx.DB, build *common.Build) error {

	var (
		startedAt time.Time
	)

	if err := connect.Get(&startedAt, "SELECT CURRENT_TIMESTAMP"); err != nil {

		return fmt.Errorf("Could not retrieve startup parameters: %v", err)
	}

	log.Debugf("Custom runner. Image: %s", image)

	log.Debugf("Create container: %s", image)

	container, err := dockerClient.CreateConteiner(image, docker.CreateContainerOptions{
		Env: build.Config.Env,
		Binds: []string{
			fmt.Sprintf("%s:%s", build.WorkingDir, WorkingDir),
		},
		Entrypoint: build.Config.Entrypoint,
		WorkingDir: WorkingDir,
	})

	if err != nil {

		return fmt.Errorf("Could not create container: %v", err)
	}

	defer container.Destroy()

	for _, command := range build.Config.Commands {

		log.Debugf("Run cmd: %s", command)

		if err := container.RunCmd(command); err != nil {

			return fmt.Errorf("Execute failed. Cmd: %s, output: %s", command, container.Output())
		}
	}

	var testsWithErrors []string

	for _, test := range build.Config.Tests {

		log.Debugf("Run test: %s", test)

		if err := container.RunCmd(test); err != nil {

			log.Debugf("Test failed: %s, output: %s", test, container.Output())

			testsWithErrors = append(testsWithErrors, test)
		}
	}

	log.Debugf("Output: %s", container.Output())

	_, err = connect.Exec(`SELECT build.add_part($1, $2, $3, $4, $5, $6, $7)`,
		build.BuildID,
		"",
		image,
		container.ID(),
		container.Output(),
		startedAt,
		"[]",
	)

	if err != nil {

		return fmt.Errorf("Could not commit a part of the build: %v", err)
	}

	if len(testsWithErrors) > 0 {

		return fmt.Errorf("Test(s) failed: %s", strings.Join(testsWithErrors, ", "))
	}

	return nil
}
