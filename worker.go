package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/postgres-ci/worker/src/app"
	"github.com/postgres-ci/worker/src/common"

	"flag"
	"fmt"
	"os"
)

var (
	debug        bool
	pathToConfig string
)

const usage = `
Postgres-CI worker

Usage:
    -c /path/to/config.yaml (if not set, then worker will use environment variables)
    -debug (enable debug mode)

Environment variables:

    ASSETS      - worker assets
    WORKING_DIR - shared as volume between worker and running conteiners
    NUM_WORKERS - number or auto (defines the number of parallel build process)
    LOG_LEVEL   - one of: info/warning/error

    == PostgreSQL server credentials

    DB_HOST
    DB_PORT
    DB_USERNAME
    DB_PASSWORD
    DB_DATABASE 

    == Docker daemon credentials

    DOCKER_ENDPOINT  - for connect to docker API (unix:///var/run/docker.sock)
    DOCKER_TLS_CERT_PATH
    DOCKER_AUTH_PASSWORD
    DOCKER_AUTH_EMAIL
    DOCKER_AUTH_SERVER_ADDRESS
`

func main() {

	flag.BoolVar(&debug, "debug", false, "")
	flag.StringVar(&pathToConfig, "c", "", "")

	flag.Usage = func() {

		fmt.Println(usage)

		os.Exit(0)
	}

	flag.Parse()

	if log.IsTerminal() {

		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05 MST",
		})

	} else {

		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05 MST",
		})
	}

	config, err := common.ReadConfig(pathToConfig)

	if err != nil {

		log.Fatalf("Error reading configuration file: %v", err)
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(config.LogLevel())
	}

	app := app.New(config)

	if debug {

		app.SetDebugMode()
	}

	app.Run()
}
