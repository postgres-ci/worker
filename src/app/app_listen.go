package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/lib/pq"

	"encoding/json"
	"time"
)

const (
	minReconnectInterval = time.Second
	maxReconnectInterval = time.Second * 5
	channel              = "postgres-ci"
)

type Task struct {
	BuildID   int64     `json:"build_id"   db:"build_id"`
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
		check  = time.Tick(time.Minute / 10)
		events = listener.NotificationChannel()
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

			a.tasks <- task

		case <-check:

			log.Debug("Checking for new tasks")

			for {

				var task Task

				if err := a.connect.Get(&task, "SELECT build_id, created_at FROM postgres_ci.get_new_task()"); err != nil {

					if e, ok := err.(*pq.Error); !ok || e.Message != "NO_NEW_TASKS" {

						log.Errorf("Could not fetch new task: %v", err)
					}

					break
				}

				a.tasks <- task
			}
		}
	}
}
