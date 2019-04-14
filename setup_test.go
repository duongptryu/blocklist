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
	} {
		c := caddy.NewTestController("dns", cfg)
		if err := setup(c); err == nil {
			t.Errorf("Expected errors, but got %v from %q", err, cfg)
		}
	}
}
