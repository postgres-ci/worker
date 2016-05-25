package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/postgres-ci/worker/src/app/cmd/runner"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"
)

func Build(assets string, docker docker.Client, connect *sqlx.DB) *build {

	return &build{
		assets:  assets,
		docker:  docker,
		connect: connect,
	}
}

type build struct {
	assets  string
	docker  docker.Client
	connect *sqlx.DB
}

func (b *build) Run(build *common.Build) error {

	fn := runner.PLpgSQL

	if len(build.Config.Tests) > 0 {

		fn = runner.Custom
	}

	for _, image := range build.Config.Images {

		if err := fn(image, b.assets, b.docker, b.connect, build); err != nil {

			log.Error(err.Error())

			return err
		}
	}

	return nil
}
