// Package xpprof copy from net/http/pprof  go 1.20.11
package xpprof

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
)

func AsyncListenAddress(addr ...string) *http.Server {
	s := InitAddress(addr...)
	go s.ListenAndServe()
	return s
}

func InitAddress(addr ...string) *http.Server {
	s := &http.Server{}
	if len(addr) > 0 {
		s.Addr = addr[0]
	}
	var mux http.ServeMux
	InitMux(&mux)
	s.Handler = &mux
	return s
}

func InitMux(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func InitCollector(name string, mux *http.ServeMux) {
	pc := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		PidFn: func() (int, error) {
			return os.Getpid(), nil
		},
		Namespace:    name,
		ReportErrors: false,
	})
	gc := collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(collectors.MetricsAll),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(pc, gc)
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.Default(),
		ErrorHandling: promhttp.ContinueOnError,
	})
	mux.Handle("/metrics", handler)
}
