package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/postgres-ci/worker/src/app/cmd"
	"github.com/postgres-ci/worker/src/common"
	"github.com/postgres-ci/worker/src/docker"

	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	MaxOpenConns    = 5
	MaxIdleConns    = 2
	ConnMaxLifetime = time.Minute
)

type command interface {
	Run(*common.Build) error
}

func New(config common.Config) *app {

	log.Debugf("Connect to postgresql server. DSN(%s)", config.Connect.DSN())

	connect, err := sqlx.Connect("postgres", config.Connect.DSN())

	if err != nil {

		log.Fatalf("Could not connect to database server '%v'", err)
	}

	dockerClient, err := docker.Bind(config.Docker)

	if err != nil {

		log.Fatalf("Could not bind to docker daemon '%v'", err)
	}

	connect.SetMaxOpenConns(MaxOpenConns)
	connect.SetMaxIdleConns(MaxIdleConns)
	connect.SetConnMaxLifetime(ConnMaxLifetime)

	app := app{
		config:  config,
		connect: connect,
		docker:  dockerClient,
		tasks:   make(chan Task),
		commands: []command{
			&cmd.Checkout{
				WorkingDir: config.WorkingDir,
			},
			&cmd.Clear{},
		},
	}

	go app.execute()

	return &app
}

type app struct {
	config   common.Config
	connect  *sqlx.DB
	docker   docker.Client
	tasks    chan Task
	commands []command
	debug    bool
}

func (a *app) SetDebugMode() {

	a.debug = true
}

func (a *app) Run() {

	log.Info("Postgres-CI worker started")
	log.Debugf("Pid: %d", os.Getpid())

	if a.debug {

		go a.debugInfo()
	}

	go a.listen()

	a.handleOsSignals()
}

func (a *app) handleOsSignals() {

	signalChan := make(chan os.Signal)

	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGUSR1,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGHUP,
	)

	for {

		switch sig := <-signalChan; sig {

		case syscall.SIGUSR1:

			log.Info("Signal 'USR1'. Logrotate (postrotate)")

		default:

			log.Infof("Postgres-CI worker stopped (signal: %v)", sig)

			os.Exit(0)
		}
	}
}
