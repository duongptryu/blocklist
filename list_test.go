package blocklist

import (
	"strings"
	"testing"
	"time"
)

func TestListParse(t *testing.T) {
	var list = `
# 127.0.0.1	example.com
127.0.0.1	example.org	third
008.free-counter.co.uk
com
`

	r := strings.NewReader(list)
	l, _ := listRead(r)
	db := NewMemoryDB()
	db.Update("test", time.Now(), l)
	db.update(db.combine())

	tests := []struct {
		name    string
		blocked bool
	}{
		{"example.org.", false},
		{"example.com.", false},
		{"com.", false},
	}

	for _, test := range tests {
		got := blocked(db, test.name)
		if got != test.blocked {
			t.Errorf("Expected %s to be blocked", test.name)
		}
	}
}
