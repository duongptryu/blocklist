package blocklist

// HashDB is a hash-backed static list of blocked domains.
type HashDB map[string]bool

func (h HashDB) Block(domain string) bool { return h[domain] }
