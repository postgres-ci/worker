package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"

	"database/sql"
	"fmt"
	"time"
)

func Build(assets string, docker docker.Client, connect *sqlx.DB) *build {

	return &build{
		assets:  assets,
		docker:  docker,
		connect: connect,
	}
}

const (
	WorkingDir = "/opt/postgres-ci/build/"
	AssetsDir  = "/opt/postgres-ci/assets/"
)

type build struct {
	assets  string
	docker  docker.Client
	connect *sqlx.DB
}

func (b *build) Run(build *common.Build) error {

	for _, image := range build.Config.Images {

		if err := b.subbuild(image, build); err != nil {

			return err
		}
	}

	return nil
}

func (b *build) subbuild(image string, build *common.Build) error {

	container, err := b.docker.CreateConteiner(image, docker.CreateContainerOptions{
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_DB=postgres",

			fmt.Sprintf("TEST_DATABASE=%s", build.Config.Postgres.Database),
			fmt.Sprintf("TEST_USERNAME=%s", build.Config.Postgres.Username),
			fmt.Sprintf("TEST_PASSWORD=%s", build.Config.Postgres.Password),
		},
		Binds: []string{
			fmt.Sprintf("%s:%s", build.WorkingDir, WorkingDir),
			fmt.Sprintf("%s:%s", b.assets, AssetsDir),
		},
		WorkingDir: WorkingDir,
	})

	if err != nil {

		log.Errorf("Could not create container: %v", err)

		return err
	}

	log.Debugf("Create container: %s", image)

	defer container.Destroy()

	if err := waitingForStartup(container.IPAddress); err != nil {

		log.Errorf("Could not run PostgreSQL server: %v", err)

		return err
	}

	for _, command := range append([]string{"bash /opt/postgres-ci/assets/setup.sh"}, build.Config.Commands...) {

		if err := container.RunCmd(command); err != nil {

			log.Errorf("Execute failed. Cmd: %s, output: %s", command, container.Output.String())

			return err
		}
	}

	return nil
}

func waitingForStartup(ipAddress string) error {

	connect, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", ipAddress))

	if err != nil {

		return err
	}

	defer connect.Close()

	for i := 0; i < 30; i++ {

		time.Sleep(time.Second)

		if err := connect.Ping(); err == nil {

			return nil
		}
	}

	return fmt.Errorf("Could not connect to: %s", ipAddress)
}
