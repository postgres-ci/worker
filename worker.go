package main

import (
	logger "github.com/Sirupsen/logrus"
	_ "github.com/lib/pq"
	"github.com/postgres-ci/worker/src/app"
	"github.com/postgres-ci/worker/src/config"

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
    -c /path/to/config.yaml (default is /etc/postgres-ci/worker.yaml)
    -debug (enable debug mode)
`

func main() {

	flag.BoolVar(&debug, "debug", false, "")
	flag.StringVar(&pathToConfig, "c", "/etc/postgres-ci/worker.yaml", "")

	flag.Usage = func() {

		fmt.Println(usage)

		os.Exit(0)
	}

	flag.Parse()

	if _, err := os.Open(pathToConfig); err != nil {

		if os.IsNotExist(err) {

			logger.Fatalf("No such configuration file '%s'", pathToConfig)
		}

		logger.Fatalf("Could not open configuration file '%s'. %v", pathToConfig, err)
	}

	config, err := config.Open(pathToConfig)

	if err != nil {

		logger.Fatalf("Error reading configuration file '%v'", err)
	}

	if debug {
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(config.Logger.LogLevel())
	}

	app := app.New(config)

	if debug {

		app.SetDebugMode()
	}

	app.Run()
}
