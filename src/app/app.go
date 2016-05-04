package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/postgres-ci/worker/src/app/cmd"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"

	"os"
	"time"
)

const (
	MaxOpenConns    = 5
	MaxIdleConns    = 2
	ConnMaxLifetime = time.Hour
)

type command interface {
	Run(*common.Build) error
}

func New(config common.Config) *app {

	log.Debugf("Connect to postgresql server. DSN(%s)", config.Connect.DSN())

	connect, err := sqlx.Connect("postgres", config.Connect.DSN())

	if err != nil {

		log.Fatalf("Could not connect to database server: %v", err)
	}

	dockerClient, err := docker.Bind(config.Docker)

	if err != nil {

		log.Fatalf("Could not bind to docker daemon: %v", err)
	}

	connect.SetMaxOpenConns(MaxOpenConns)
	connect.SetMaxIdleConns(MaxIdleConns)
	connect.SetConnMaxLifetime(ConnMaxLifetime)

	app := app{
		config:  config,
		connect: connect,
		commands: []command{
			&cmd.Checkout{
				WorkingDir: config.WorkingDir,
			},
			cmd.Build(
				config.Assets,
				dockerClient,
				connect,
			),
			&cmd.Clear{},
		},
		tasks: make(chan Task),
	}

	for i := 0; i < config.GetNumWorkers(); i++ {

		go app.worker()
	}

	return &app
}

type app struct {
	config   common.Config
	connect  *sqlx.DB
	commands []command
	tasks    chan Task
	debug    bool
}

func (a *app) SetDebugMode() {

	a.debug = true
}

func (a *app) Run() {

	log.Info("Postgres-CI worker started")
	log.Debugf("Pid: %d, num-workers: %d", os.Getpid(), a.config.GetNumWorkers())

	if a.debug {

		go a.debugInfo()
	}

	a.listen()
}
