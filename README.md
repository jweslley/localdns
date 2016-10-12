# localdns

[![Travis](https://api.travis-ci.org/jweslley/localdns.png)](http://travis-ci.org/jweslley/localdns)
[![Go Report Card](https://goreportcard.com/badge/github.com/jweslley/localdns)](https://goreportcard.com/report/github.com/jweslley/localdns)

A DNS for local development. Editing hosts file to add and remove host names for local development is no longer needed.

localdns is designed to respond to DNS queries for all subdomains of the specified top-level domain with localhost address. Since it supports both IPv4 and IPv6, localdns will respond accordingly to, i.e., it will respond to DNS `A` queries with `127.0.0.1` and `AAAA` queries with `::1`.

localdns also can be used as a DNS proxy. This feature is specially useful in Windows since default Windows DNS's client [does not query a secondary DNS](https://groups.google.com/forum/#!topic/microsoft.public.windows.server.active_directory/wcNs42YNKeo) before a [15 minute timeout](https://support.microsoft.com/en-us/kb/320760/en-us?p=1).

Supports Linux, Mac OSX and Windows! \o/

## Installation

### General

[Download](https://github.com/jweslley/localdns/releases) and put the binary somewhere in your path.

### Archlinux (AUR package)

    yaourt -S localdns

> Installing using `yaourt` also creates a systemd service: `localdns.service`.

### Mac OSX (Homebrew)

    brew tap jweslley/formulae
    brew install localdns

> Installing using brew creates a [plist](https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/plist.5.html) file to launches localdns via launchd and create a custom [resolver](https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/resolver.5.html) to `.dev` top-level domains.

### From source

    git clone git://github.com/jweslley/localdns.git
    cd localdns
    make build

## Usage

After installed, running localdns is straightforward, just run:

    localdns

For more command options, run `localdns -h`:

    -tld="dev": Top-level domain to resolve to localhost
    -port=5353: DNS's port
    -ttl=600: DNS's TTL (Time to live)
    -debug: enable verbose logging

    -proxy: Enable proxy mode. comma-separated list of DNS servers to send queries for. Example: 8.8.8.8,8.8.4.4
    -timeout=2s: when acting as proxy, timeout for dial, write and read.
    -expire=10m: when acting as proxy, cache expiration time.
    -cache=65536: when acting as proxy, the cache size.

In Windows, is recommended to enable the proxy mode. For example, run:

    localdns -proxy 8.8.8.8,8.8.4.4


## Test usage

You can use `dig`, `drill` or other tool to run queries against your localdns instance.

Executing a query using `drill`:

    drill @localhost -p 5353 myapp.dev

Outputs:

    ;; ->>HEADER<<- opcode: QUERY, rcode: NOERROR, id: 21160
    ;; flags: qr rd ; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
    ;; QUESTION SECTION:
    ;; myapp.dev.   IN      A

    ;; ANSWER SECTION:
    myapp.dev.      600     IN      A       127.0.0.1

    ;; AUTHORITY SECTION:

    ;; ADDITIONAL SECTION:

    ;; Query time: 0 msec
    ;; SERVER: ::1
    ;; WHEN: Thu May 14 12:44:43 2015
    ;; MSG SIZE  rcvd: 52


## Bugs and Feedback

If you discover any bugs or have some idea, feel free to create an issue on GitHub:

    http://github.com/jweslley/localdns/issues


## License

MIT license. Copyright (c) 2015 Jonhnny Weslley <http://jonhnnyweslley.net>

See the LICENSE file provided with the source distribution for full details.
