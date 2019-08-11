// Package blocklist contains a blocklist plugin for CoreDNS.
package blocklist

import (
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
	db          *MemoryDB
	manualAllow HashDB
	manualBlock HashDB
	lists       map[string]*List
	stop, poke  chan struct{}

	Next plugin.Handler
}

// New returns a new Blocklist.
func New(db *MemoryDB) *Blocklist {
	return &Blocklist{
		db:          db,
		manualAllow: make(HashDB),
		manualBlock: make(HashDB),
		lists:       make(map[string]*List),
		poke:        make(chan struct{}, 1),
	}
}

// Start starts the internals of Blocklist.
func (b *Blocklist) Start() error {
	b.stop = make(chan struct{})
	go b.db.Pokee(b.stop, b.poke)
	for _, v := range b.lists {
		go v.Run(b.db, b.stop, b.poke)
	}
	return nil
}

// Stop stops the internals of Blocklist.
func (b *Blocklist) Stop() error {
	close(b.stop)
	b.stop = nil
	return nil
}

func (b *Blocklist) isBlocked(name string) bool {
	switch {
	case blocked(b.manualAllow, name):
		return false
	case blocked(b.manualBlock, name):
		return true
	case blocked(b.db, name):
		return true
	default:
		return false
	}
}

// ServeDNS implements the plugin.Handler interface.
func (b *Blocklist) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	block := b.isBlocked(state.Name())
	if block {
		blockCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

		resp := new(dns.Msg)
		resp.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(resp)

		return dns.RcodeNameError, nil
	}

	return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
}

// Name implements the Handler interface.
func (b *Blocklist) Name() string { return "blocklist" }

// Blocker returns true for blocked DNS domains.
type Blocker interface {
	Block(string) bool
}

// blocked returns true when name is in list or is a subdomain for any names in the list. "localhost." is never blocked.
func blocked(db Blocker, name string) bool {
	if name == "localhost." {
		return false
	}
	blocked := db.Block(name)
	if blocked {
		return true
	}
	i, end := dns.NextLabel(name, 0)
	for !end {
		blocked := db.Block(name[i:])
		if blocked {
			return true
		}
		i, end = dns.NextLabel(name, i)
	}
	return false
}
