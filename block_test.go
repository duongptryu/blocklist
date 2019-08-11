package blocklist

import (
	"strings"
	"testing"
	"time"
)

func newTestDB() *MemoryDB {
	var list = `
	127.0.0.1	005.free-counter.co.uk
	127.0.0.1	006.free-adult-counters.x-xtra.com
	127.0.0.1	006.free-counter.co.uk
	127.0.0.1	007.free-counter.co.uk
	127.0.0.1	007.go2cloud.org
	127.0.0.1	localhost
	008.free-counter.co.uk
	com
	`

	r := strings.NewReader(list)
	l, _ := listRead(r)
	db := NewMemoryDB()
	db.Update("test", time.Now(), l)
	db.update(db.combine())
	return db
}

func TestBlocked(t *testing.T) {
	db := newTestDB()

	tests := []struct {
		name    string
		blocked bool
	}{
		{"example.org.", false},
		{"localhost.", false},
		{"com.", false},

		{"005.free-counter.co.uk.", true},
		{"www.005.free-counter.co.uk.", true},
		{"008.free-counter.co.uk.", true},
		{"www.008.free-counter.co.uk.", true},
	}

	for _, test := range tests {
		got := blocked(db, test.name)
		if got != test.blocked {
			t.Errorf("Expected %s to be blocked", test.name)
		}
	}
}

func TestBlockHierarchy(t *testing.T) {
	b := New(newTestDB())

	// allow subdomain of blocked domain
	b.manualAllow["subdomainonly.005.free-counter.co.uk."] = true
	// allow list-blocked domain
	b.manualAllow["007.free-counter.co.uk."] = true
	// allow non-blocked domain
	b.manualAllow["distinctdomain.com."] = true
	// allow manually-blocked domain
	b.manualAllow["alsoblocked.com."] = true
	b.manualBlock["alsoblocked.com."] = true
	// block list-blocked domain
	b.manualBlock["007.go2cloud.org."] = true
	// block non-blocked domain
	b.manualBlock["onlyblocked.com."] = true

	tests := []struct {
		name    string
		blocked bool
	}{
		{"005.free-counter.co.uk.", true},
		{"www.005.free-counter.co.uk.", true},
		{"subdomainonly.005.free-counter.co.uk.", false},
		{"subsub.subdomainonly.005.free-counter.co.uk.", false},
		{"007.free-counter.co.uk.", false},
		{"www.007.free-counter.co.uk.", false},
		{"distinctdomain.com.", false},
		{"sub.distinctdomain.com.", false},
		{"alsoblocked.com.", false},
		{"sub.alsoblocked.com.", false},
		{"007.go2cloud.org.", true},
		{"onlyblocked.com.", true},
		{"sub.onlyblocked.com.", true},
	}

	for _, test := range tests {
		got := b.isBlocked(test.name)
		if got != test.blocked {
			want := "unblocked"
			if test.blocked {
				want = "blocked"
			}
			t.Errorf("Expected %s to be %s", test.name, want)
		}
	}
}
