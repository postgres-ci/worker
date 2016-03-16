package app

import (
	logger "github.com/Sirupsen/logrus"
	"github.com/lib/pq"

	"time"
)

const (
	minReconnectInterval = time.Second
	maxReconnectInterval = time.Second * 5
	channel              = "postgres-ci"
)

func (a *app) listen() {

	listener := pq.NewListener(a.config.Connect.DSN(), minReconnectInterval, maxReconnectInterval, func(event pq.ListenerEventType, err error) {

		logger.Debugf("pg notify send event [%v], err '%v'", event, err)
	})

	logger.Infof("LISTEN '%s' channel", channel)

	listener.Listen(channel)

	notifications := listener.NotificationChannel()

	for {

		notification := <-notifications

		if notification == nil {
			continue
		}

		logger.Debugf("received from [%s] playload: %s", notification.Channel, notification.Extra)
	}
}
