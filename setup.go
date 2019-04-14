package blocklist

import (
	"fmt"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("blocklist", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	block, err := blocklistParse(c)
	if err != nil {
		return plugin.Error("blocklist", err)
	}

	c.OnStartup(func() error {
		metricSetup(c)
		go func() { block.download() }()
		go func() { block.refresh() }()
		return nil
	})

	c.OnShutdown(func() error {
		close(block.stop)
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		block.Next = next
		return block
	})

	return nil
}

func blocklistParse(c *caddy.Controller) (*Blocklist, error) {
	for c.Next() {
		var url string
		if c.NextArg() {
			url = c.Val()
		} else {
			return nil, c.ArgErr()
		}
		if a := c.RemainingArgs(); len(a) > 0 {
			return nil, c.SyntaxErr("each blockfile directive takes only one URL argument")
		}
		for c.NextBlock() {

		}
		fmt.Printf("parsed %q\n", url)
	}
	return nil, nil
}
