package runner

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"

	"fmt"
	"time"
)

func PLpgSQL(image, assets string, dockerClient docker.Client, connect *sqlx.DB, build *common.Build) error {

	var (
		startedAt     time.Time
		serverVersion string
	)

	if err := connect.Get(&startedAt, "SELECT CURRENT_TIMESTAMP"); err != nil {

		return fmt.Errorf("Could not retrieve startup parameters: %v", err)
	}

	log.Debugf("Create container: %s", image)

	container, err := dockerClient.CreateConteiner(image, docker.CreateContainerOptions{
		Env: append(build.Config.Env, []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_DB=postgres",

			fmt.Sprintf("TEST_DATABASE=%s", build.Config.Postgres.Database),
			fmt.Sprintf("TEST_USERNAME=%s", build.Config.Postgres.Username),
			fmt.Sprintf("TEST_PASSWORD=%s", build.Config.Postgres.Password),
		}...),
		Binds: []string{
			fmt.Sprintf("%s:%s", build.WorkingDir, WorkingDir),
			fmt.Sprintf("%s:%s", assets, AssetsDir),
		},
		Entrypoint: build.Config.Entrypoint,
		WorkingDir: WorkingDir,
	})

	if err != nil {

		return fmt.Errorf("Could not create container: %v", err)
	}

	defer container.Destroy()

	serverVersion, err = waitingForStartup(container.IPAddress)

	if err != nil {

		return fmt.Errorf("Could not run PostgreSQL server: %v", err)
	}

	log.Debugf("PostgreSQL %s database server started", serverVersion)

	for _, command := range append([]string{"bash /opt/postgres-ci/assets/setup.sh"}, build.Config.Commands...) {

		log.Debugf("Run cmd: %s", command)

		if err := container.RunCmd(command); err != nil {

			return fmt.Errorf("Execute failed. Cmd: %s, output: %s", command, container.Output())
		}
	}

	log.Debugf("Output: %s", container.Output())

	connectToContainer, err := sqlx.Connect("postgres", build.DSN(container.IPAddress))

	if err != nil {

		return fmt.Errorf("Could not connect to PostgreSQL server: %v", err)
	}

	defer connectToContainer.Close()

	var (
		tests plpgsqlTests
		sql   = `
		SELECT 
			namespace,
			procedure,
			to_json(errors) AS errors,
			started_at,
			finished_at
		FROM assert.test_runner()
		`
	)
	if err := connectToContainer.Select(&tests, sql); err != nil {

		return fmt.Errorf("Error when running a tests: %v", err)
	}

	for _, test := range tests {

		if len(test.Errors) != 0 {

			log.Debugf("--- FAIL: %s.%s %v (%.4fs)\n\t", test.Namespace, test.Procedure, test.Errors, test.FinishedAt.Sub(test.StartedAt).Seconds())

		} else {

			log.Debugf("--- PASS: %s.%s (%.4fs)", test.Namespace, test.Procedure, test.FinishedAt.Sub(test.StartedAt).Seconds())
		}
	}

	_, err = connect.Exec(`SELECT build.add_part($1, $2, $3, $4, $5, $6, $7)`,
		build.BuildID,
		serverVersion,
		image,
		container.ID(),
		container.Output(),
		startedAt,
		tests,
	)

	if err != nil {

		return fmt.Errorf("Could not commit a part of the build: %v", err)
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

		time.Sleep(time.Second * 2)

		if err := connect.Get(&serverVersion, "SHOW server_version"); err == nil {

			return serverVersion, nil
		}
	}

	return "", fmt.Errorf("connect timeout: %s", ipAddress)
}
