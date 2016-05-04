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
