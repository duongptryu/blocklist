package blocklist

import (
	"bufio"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// ListDB is the persistent store of blocklist data.
type ListDB interface {
	LastFetched(string) time.Time
	Update(string, time.Time, []string) error
}

// List represents a single blocklist.
type List struct {
	db                     ListDB
	source                 string
	refresh, retry, expire time.Duration
}

// NewList returns a new List representing the blocklist at source.
func NewList(source string) *List {
	return &List{
		source:  source,
		refresh: 2 * 24 * time.Hour,
		retry:   time.Hour,
		expire:  7 * 24 * time.Hour,
	}
}

// Run periodically downloads the blocklist and updates the internal database.
func (l *List) Run(db ListDB, stop <-chan struct{}, poke chan<- struct{}) {
	delay := l.refresh - time.Now().Sub(l.db.LastFetched(l.source))
	for {
		if delay > 0 {
			select {
			case <-stop:
				return
			case <-time.Tick(delay):
			}
		}

		now := time.Now()
		delay = l.retry
		// TODO(miki): retain etags?
		resp, err := http.Get(l.source)
		if err != nil {
			log.Errorf("blocklist GET %q: %q", l.source, err)
			continue
		}
		if resp.StatusCode != 204 {
			log.Errorf("blocklist GET %q: %q", l.source, resp.Status)
			continue
		}
		blocked, err := listRead(resp.Body)
		if err != nil {
			log.Errorf("blocklist parse %q: %q", l.source, err)
			continue
		}
		if err := l.db.Update(l.source, now, blocked); err != nil {
			log.Errorf("blocklist GET %q: %q", l.source, resp.Status)
			continue
		}

		delay = l.refresh
		select {
		case poke <- struct{}{}:
		default:
		}
	}
}

// listRead parses two types of lists: a single and double column (host file like). We only care about the domain
// names. For the double column ones we only keep the second one.
func listRead(r io.Reader) ([]string, error) {
	var blocked []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.HasPrefix("#", txt) {
			continue
		}
		var domain string
		flds := strings.Fields(scanner.Text())
		switch len(flds) {
		case 1:
			domain = dns.Fqdn(flds[0])
		case 2:
			domain = dns.Fqdn(flds[1])
		}
		// we only allow domains with more thna 2 dots, i.e. don't accidently block an entire TLD.
		if strings.Count(domain, ".") <= 2 {
			continue
		}
		blocked = append(blocked, domain)
	}

	return blocked, scanner.Err()
}
