package blocklist

import (
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var log = clog.NewWithPlugin("blocklist")

// Blocklist is the blocklist plugin.
type Blocklist struct {
	list map[string]struct{}

	update map[string]struct{}
	sync.RWMutex
	stop chan struct{}

	Next plugin.Handler
}

func New() *Blocklist {
	return &Blocklist{
		list:   make(map[string]struct{}),
		update: make(map[string]struct{}),
		stop:   make(chan struct{}),
	}
}

// ServeDNS implements the plugin.Handler interface.
func (b *Blocklist) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	if b.blocked(state.Name()) {
		blockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
		log.Infof("Blocked %s", state.Name())

		resp := new(dns.Msg)
		resp.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(resp)

		return dns.RcodeNameError, nil
	}

	return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (b *Blocklist) Name() string { return "blocklist" }

// blocked returns true when name is in list or is a subdomain for any names in the list. "localhost." is never blocked.
func (b *Blocklist) blocked(name string) bool {
	b.RLock()
	defer b.RUnlock()

	if name == "localhost." {
		return false
	}
	_, blocked := b.list[name]
	if blocked {
		return true
	}
	i, end := dns.NextLabel(name, 0)
	for !end {
		_, blocked := b.list[name[i:]]
		if blocked {
			return true
		}
		i, end = dns.NextLabel(name, i)
	}
	return false
}
