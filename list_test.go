package blocklist

import (
	"strings"
	"testing"
)

func TestListParse(t *testing.T) {
	var list = `
# 127.0.0.1	example.com
127.0.0.1	example.org	third
008.free-counter.co.uk
com
`

	b := new(Blocklist)
	r := strings.NewReader(list)
	l := make(map[string]struct{})
	listRead(r, l)

	b.list = l

	tests := []struct {
		name    string
		blocked bool
	}{
		{"example.org.", false},
		{"example.com.", false},
		{"com.", false},
	}

	for _, test := range tests {
		got := b.blocked(test.name)
		if got != test.blocked {
			t.Errorf("Expected %s to be blocked", test.name)
		}
	}
}
