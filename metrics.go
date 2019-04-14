package blocklist

import (
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	"github.com/mholt/caddy"

	"github.com/prometheus/client_golang/prometheus"
)

var blockCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "blocklist",
	Name:      "count_total",
	Help:      "Counter of blocked names.",
}, []string{"server"})

var once sync.Once

func metricSetup(c *caddy.Controller) {
	once.Do(func() { metrics.MustRegister(c, blockCount) })
}
