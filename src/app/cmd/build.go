package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"

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

		if err := b.runPart(image, build); err != nil {

			return err
		}
	}

	return nil
}

const (
	TestRunnerSql = `
		SELECT 
			namespace,
			procedure,
			to_json(errors) AS errors,
			started_at,
			finished_at
		FROM assert.test_runner()
	`
)

func (b *build) runPart(image string, build *common.Build) error {

	var (
		startedAt     time.Time
		serverVersion string
	)

	if err := b.connect.Get(&startedAt, "SELECT CURRENT_TIMESTAMP"); err != nil {

		log.Errorf("Could not retrieve startup parameters: %v", err)

		return err
	}

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

	serverVersion, err = waitingForStartup(container.IPAddress)

	if err != nil {

		log.Errorf("Could not run PostgreSQL server: %v", err)

		return err
	}

	log.Debugf("PostgreSQL %s database server started", serverVersion)

	for _, command := range append([]string{"bash /opt/postgres-ci/assets/setup.sh"}, build.Config.Commands...) {

		log.Debugf("Run cmd: %s", command)

		if err := container.RunCmd(command); err != nil {

			log.Errorf("Execute failed. Cmd: %s, output: %s", command, container.Output())

			return err
		}
	}

	log.Debugf("Output: %s", container.Output())

	connect, err := sqlx.Connect("postgres", build.DSN(container.IPAddress))

	if err != nil {

		log.Errorf("Could not connect to PostgreSQL server: %v", err)

		return err
	}

	defer connect.Close()

	var tests plpgsqlTests

	if err := connect.Select(&tests, TestRunnerSql); err != nil {

		log.Errorf("Error when running a tests: %v", err)

		return err
	}

	for _, test := range tests {

		if len(test.Errors) != 0 {

			log.Debugf("--- FAIL: %s.%s %v (%.4fs)\n\t", test.Namespace, test.Procedure, test.Errors, test.FinishedAt.Sub(test.StartedAt).Seconds())

		} else {

			log.Debugf("--- PASS: %s.%s (%.4fs)", test.Namespace, test.Procedure, test.FinishedAt.Sub(test.StartedAt).Seconds())
		}
	}

	_, err = b.connect.Exec(`SELECT build.add_part($1, $2, $3, $4, $5, $6, $7)`,
		build.BuildID,
		serverVersion,
		image,
		container.ID(),
		container.Output(),
		startedAt,
		tests,
	)

	if err != nil {

		log.Errorf("Could not commit a part of the build: %v", err)

		return err
	}

	return nil
}

func waitingForStartup(ipAddress string) (string, error) {

	connect, err := sqlx.Open("postgres", fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", ipAddress))

	if err != nil {

		return "", err
	}

	defer connect.Close()

	var serverVersion string

	for i := 0; i < 30; i++ {

		time.Sleep(time.Second)

		if err := connect.Get(&serverVersion, "SHOW server_version"); err == nil {

			return serverVersion, nil
		}
	}

	return "", fmt.Errorf("Could not connect to: %s", ipAddress)
}
