# localdns

[![Travis](https://api.travis-ci.org/jweslley/localdns.png)](http://travis-ci.org/jweslley/localdns)

A DNS for local development. Editing hosts file to add and remove host names for local development is no longer needed.

localdns is designed to respond to DNS queries for all subdomains of the specified top-level domain with localhost address. Since it supports both IPv4 and IPv6, localdns will respond accordingly to, i.e., it will respond to DNS `A` queries with `127.0.0.1` and `AAAA` queries with `::1`.


## Installation

[Download](https://github.com/jweslley/localdns/releases) and put the binary somewhere in your path.

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
