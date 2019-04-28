package blocklist

import (
	"sync"
	"time"
)

// MemoryDB is an in-memory store of blocklist data.
type MemoryDB struct {
	mu          sync.RWMutex
	blocked     map[string]bool
	lastFetched map[string]time.Time
	lists       map[string][]string
}

// NewMemoryDB returns a new MemoryDB. A client must call Pokee on a separate goroutine.
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		blocked:     make(map[string]bool),
		lastFetched: make(map[string]time.Time),
		lists:       make(map[string][]string),
	}
}

// LastFetched returns the time that the given source was last fetched, or the zero time if it has never been fetched.
func (db *MemoryDB) LastFetched(source string) time.Time {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.lastFetched[source]
}

// Update sets the contents of the source to blocked as of time fetched.
func (db *MemoryDB) Update(source string, fetched time.Time, blocked []string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.lists[source] = blocked
	db.lastFetched[source] = fetched
	return nil
}

// Blocked returns true if domain is blocked.
func (db *MemoryDB) Blocked(domain string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.blocked[domain]
}

func (db *MemoryDB) Pokee(stop, poke <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-poke:
			db.update(db.combine())
		}
	}
}

func (db *MemoryDB) combine() map[string]bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	blocked := make(map[string]bool, len(db.blocked))
	for _, list := range db.lists {
		for _, domain := range list {
			blocked[domain] = true
		}
	}
	return blocked
}

func (db *MemoryDB) update(blocked map[string]bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.blocked = blocked
}
