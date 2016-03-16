package app

import (
	logger "github.com/Sirupsen/logrus"

	"runtime"
	"time"
)

func (a *app) debugInfo() {

	tick := time.Tick(time.Minute)

	for {

		var memStats runtime.MemStats

		runtime.ReadMemStats(&memStats)

		logger.Debugf(
			"gorutines: %d, num gc: %d, alloc: %d, mallocs: %d, frees: %d, heap alloc: %d, stack inuse: %d",
			runtime.NumGoroutine(),
			memStats.NumGC,
			memStats.Alloc,
			memStats.Mallocs,
			memStats.Frees,
			memStats.HeapAlloc,
			memStats.StackInuse,
		)

		<-tick
	}
}
