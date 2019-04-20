package blocklist

import (
	"sync"
	"time"
)

type MemoryDB struct {
	mu          sync.RWMutex
	blocked     map[string]bool
	lastFetched map[string]time.Time
	lists       map[string][]string
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		blocked:     make(map[string]bool),
		lastFetched: make(map[string]time.Time),
		lists:       make(map[string][]string),
	}
}

func (db *MemoryDB) LastFetched(source string) time.Time {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.lastFetched[source]
}

func (db *MemoryDB) Update(source string, fetched time.Time, blocked []string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.lists[source] = blocked
	db.lastFetched[source] = fetched
	return nil
}

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
