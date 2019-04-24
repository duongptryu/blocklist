package blocklist

import (
	"net/url"

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

	c.OnFirstStartup(func() error {
		listMetrics(c)
		metricSetup(c)
		return nil
	})
	c.OnStartup(block.Start)
	c.OnShutdown(block.Stop)

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
	b := New(NewMemoryDB())
	for c.Next() {
		url, err := expectURLArg(c)
		if err != nil {
			return nil, err
		}
		for c.NextBlock() {
			var s string
			var err error
			switch c.Val() {
			case "always_allow":
				s, err = expectOneArg(c)
				b.manualAllow[s] = true
			case "block":
				s, err = expectOneArg(c)
				b.manualBlock[s] = true
			default:
				err = c.ArgErr()
			}
			if err != nil {
				return nil, err
			}
		}
		if url == "override" {
			continue
		}
		b.lists[url] = NewList(url)
	}
	return b, nil
}

func expectOneArg(c *caddy.Controller) (string, error) {
	if !c.NextArg() {
		return "", c.ArgErr()
	}
	ret := c.Val()
	if a := c.RemainingArgs(); len(a) > 0 {
		return "", c.SyntaxErr("only one argument on line")
	}
	return ret, nil
}

func expectURLArg(c *caddy.Controller) (string, error) {
	s, err := expectOneArg(c)
	if s == "override" {
		return s, err
	}
	u, err := url.Parse(s)
	if err != nil || !u.IsAbs() {
		return s, c.SyntaxErr(`valid URL or "override" keyword`)
	}
	return u.String(), err
}
