package blocklist

import (
	"testing"

	"github.com/mholt/caddy"
)

func TestSetupValidCfg(t *testing.T) {
	for _, cfg := range []string{
		`blocklist https://foo.bar`,
		`blocklist http://baz.wop/dir/path {
			always_allow fish
		}`,
		`blocklist file:///a {
			block b
			always_allow c
		}`,
	} {
		c := caddy.NewTestController("dns", cfg)
		if err := setup(c); err != nil {
			t.Errorf("Expected no errors, but got %v from %q", err, cfg)
		}
	}
}

func TestSetupInvalidCfg(t *testing.T) {
	for _, cfg := range []string{
		`blocklist`,
		`blocklist https://foo.bar a`,
		`blocklist https://foo.bar {
			always_allow b c
		}`,
		`blocklist https://foo.bar {
			frog b
		}`,
		`blocklist https://foo.bar {
			fish
		}`,
	} {
		c := caddy.NewTestController("dns", cfg)
		if err := setup(c); err == nil {
			t.Errorf("Expected errors, but got %v from %q", err, cfg)
		}
	}
}
