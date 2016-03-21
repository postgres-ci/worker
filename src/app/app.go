package app

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/postgres-ci/worker/src/config"
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

func New(config config.Config) app {

	logger.Debugf("Connect to postgresql server. DSN(%s)", config.Connect.DSN())

	connect, err := sqlx.Connect("postgres", config.Connect.DSN())

	if err != nil {

		logger.Fatalf("Could not connect to database server '%v'", err)
	}

	dockerClient, err := docker.Bind(config.Docker)

	if err != nil {

		logger.Fatalf("Could not bind to docker daemon '%v'", err)
	}

	connect.SetMaxOpenConns(MaxOpenConns)
	connect.SetMaxIdleConns(MaxIdleConns)
	connect.SetConnMaxLifetime(ConnMaxLifetime)

	return app{
		config:  config,
		connect: connect,
		docker:  dockerClient,
	}
}

type app struct {
	config  config.Config
	connect *sqlx.DB
	docker  docker.Client
	debug   bool
}

func (a *app) SetDebugMode() {

	a.debug = true
}

func (a *app) Run() {

	logger.Info("Postgres-CI worker started")
	logger.Debugf("Pid: %d", os.Getpid())

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
			logger.Info("Signal 'USR1'. Logrotate (postrotate)")
		default:
			logger.Infof("Postgres-CI worker stopped (signal: %v)", sig)
			os.Exit(0)
		}
	}
}
