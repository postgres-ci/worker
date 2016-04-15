package logwriter

import (
	log "github.com/Sirupsen/logrus"

	"io"
	"os"
	"sync"
)

func New(path string) *logwriter {

	logger := &logwriter{
		path:   path,
		writer: os.Stderr,
	}

	if !log.IsTerminal() {

		log.SetFormatter(&log.JSONFormatter{})

		logger.open()
	}

	return logger
}

type logwriter struct {
	path   string
	mutex  sync.Mutex
	writer io.Writer
}

func (l *logwriter) open() {

	if log.IsTerminal() {

		return
	}

	l.mutex.Lock()

	defer l.mutex.Unlock()

	file, err := os.OpenFile(l.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {

		log.Errorf("Could not open log file (err: %v)", err)

		return
	}

	if file, ok := l.writer.(*os.File); ok {

		file.Close()
	}

	l.writer = file

	log.SetOutput(l.writer)
}

func (l *logwriter) Write(p []byte) (n int, err error) {

	l.mutex.Lock()

	defer l.mutex.Unlock()

	return l.writer.Write(p)
}

func (l *logwriter) Logrotate() {

	l.open()
}
