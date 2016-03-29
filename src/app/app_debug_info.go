package app

import (
	log "github.com/Sirupsen/logrus"
	_ "github.com/mkevac/debugcharts"
	_ "net/http/pprof"

	"net/http"
	"runtime"
	"time"
)

func (a *app) debugInfo() {

	go func() {

		log.Info("Worker running in debug mode")
		log.Infof("pprof: http://%s/debug/pprof", a.config.DebugAddr())
		log.Infof("Debug charts: http://%s/debug/charts", a.config.DebugAddr())
		log.Error(http.ListenAndServe(a.config.DebugAddr(), nil))
	}()

	tick := time.Tick(time.Minute)

	for {

		var memStats runtime.MemStats

		runtime.ReadMemStats(&memStats)

		log.Debugf(
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
