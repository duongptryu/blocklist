package blocklist

import (
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

func metricSetup(c *caddy.Controller) error {
	// TODO(miki): this should return the error rather than panicing
	metrics.MustRegister(c, blockCount)
	return nil
}
