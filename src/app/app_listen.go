package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"

	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

const (
	minReconnectInterval = time.Second
	maxReconnectInterval = 5 * time.Second
	containerZombieTTL   = 2 * time.Hour
	channel              = "postgres-ci::tasks"
)

type Task struct {
	BuildID   int32     `json:"build_id"   db:"build_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (a *app) listen() {

	listener := pq.NewListener(a.config.Connect.DSN(), minReconnectInterval, maxReconnectInterval, func(event pq.ListenerEventType, err error) {

		if err != nil {

			log.Errorf("Postgres listen: %v", err)

			return
		}

		log.Debugf("Postgres notify send event: %v", event)
	})

	listener.Listen(channel)

	var (
		events          = listener.NotificationChannel()
		checkTasks      = time.Tick(time.Minute)
		checkContainers = time.Tick(time.Minute * 10)
	)

	for {

		select {

		case event := <-events:

			if event == nil {

				continue
			}

			log.Debugf("Received from [%s] playload: %s", event.Channel, event.Extra)

			var task Task

			if err := json.Unmarshal([]byte(event.Extra), &task); err != nil {

				log.Errorf("Could not unmarshal notify playload: %v", err)

				continue
			}

			var accept bool

			if err := a.connect.Get(&accept, "SELECT accept FROM build.accept($1)", task.BuildID); err == nil && accept {

				a.tasks <- task

			} else {

				log.Debugf("Error when accepting a task: %v", err)
			}

		case <-checkTasks:

			log.Debug("Checking for new tasks")

			for {

				var task Task

				if err := a.connect.Get(&task, "SELECT build_id, created_at FROM build.fetch()"); err != nil {

					if err != sql.ErrNoRows {

						log.Errorf("Could not fetch new task: %v", err)
					}

					break
				}

				a.tasks <- task
			}

		case <-checkContainers:

			if containers, err := a.docker.ListContainers(); err == nil {

				for _, container := range containers {

					if strings.HasPrefix(container.Name, "/pci-seq-") && container.CreatedAt.Add(containerZombieTTL).Before(time.Now()) {

						log.Warnf("Container was running too long time, destroy: %s", container.Name)

						container.Destroy()
					}
				}
			}
		}
	}
}
