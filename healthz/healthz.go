package healthz

import (
	"net/http"
	"time"

	"github.com/raintank/raintank-probe/scheduler"
	log "github.com/sirupsen/logrus"
)

type Healthz struct {
	server       *http.Server
	jobScheduler *scheduler.Scheduler
}

// NewHealthz runs a HTTP server, accepting requests to /ready and /alive which reports the
// readiness/liveness of the probe
func NewHealthz(jobScheduler *scheduler.Scheduler, addr string) *Healthz {
	h := Healthz{
		jobScheduler: jobScheduler,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ready", h.ReadyHandler())
	mux.HandleFunc("/alive", h.AliveHandler())
	s := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	h.server = s
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			log.Errorf("healthz server error. %s", err)
		}
	}()
	return &h
}

func (h *Healthz) Stop() {
	h.server.Close()
	log.Info("healthz server closed")
}

func (h *Healthz) ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthy := h.jobScheduler.IsHealthy()
		if healthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	}
}

func (h *Healthz) AliveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
