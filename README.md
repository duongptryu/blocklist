# block

## Name

*blocklist* - uses pi-hole-like block lists to block nefarious domains.

## Description

The blocklist plugin will fetch configured blocklists from the internet and block local clients from resolving the domains listed on them.

For a domain that is blocked we will return a NXDOMAIN response.

This plugin is a WIP.

## Syntax

~~~ txt
blocklist https://hosts-file.net/ad_servers.txt
~~~

(see also the sample Corefile in this directory)

## Metrics

If monitoring is enabled (via the *prometheus* directive) the following metric is exported:

* `coredns_blocklist_count_total{server}` - counter of total number of blocked domains.
* `coredns_blocklist_fetch{list, result}` - counter of list fetch attempts and the results of the fetch operation.
* `coredns_blocklist_list_size{list}` - number of blocked domains on each configured list.

The `list` label contains the URL of the blocklist in question; the `result` label is either `OK` or a brief error string.

The `server` label indicates which server handled the request, see the *metrics* plugin for details.
