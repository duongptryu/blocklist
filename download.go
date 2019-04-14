package blocklist

import (
	"net/http"
	"time"
)

// our default blocklists.
var blocklist = map[string]string{
	"StevenBlack": "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"MalwareDom":  "https://mirror1.malwaredomains.com/files/justdomains",
	"ZeusTracker": "https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
	"DisconAd":    "https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"HostsFile":   "https://hosts-file.net/ad_servers.txt",
}

func (b *Blocklist) download() {
	domains := 0
	for _, url := range blocklist {
		log.Infof("Blocklist update started %q", url)
		resp, err := http.Get(url)
		if err != nil {
			log.Warningf("Failed to download blocklist %q: %s", url, err)
			continue
		}
		if err := listRead(resp.Body, b.update); err != nil {
			log.Warningf("Failed to parse blocklist %q: %s", url, err)
		}
		domains += len(b.update)
		resp.Body.Close()

		log.Infof("Blocklist update finished %q", url)
	}
	b.Lock()
	b.list = b.update
	b.update = make(map[string]struct{})
	b.Unlock()

	log.Infof("Blocklists updated: %d domains added", domains)
}

func (b *Blocklist) refresh() {
	tick := time.NewTicker(1 * week)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			b.download()

		case <-b.stop:
			return
		}
	}
}

const week = time.Hour * 24 * 7
