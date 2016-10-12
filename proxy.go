package main

import (
	"crypto/md5"
	"fmt"
	"strings"
	"sync"
	"time"

	"net"

	"github.com/bluele/gcache"
	"github.com/miekg/dns"
)

const defaultDNSPort = "53"

type Proxy struct {
	nameservers []string

	udp *dns.Client
	tcp *dns.Client

	cache    gcache.Cache
	negCache gcache.Cache
}

func (p *Proxy) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	q, _ := questionFor(r)

	key := keyFor(q)
	if q != nil {
		Debug("%s lookup %s\n", w.RemoteAddr(), q)

		value, err := p.cache.Get(key)
		if err == nil {
			Debug("%s hit cache", q)
			msg := *(value.(*dns.Msg))
			msg.Id = r.Id
			w.WriteMsg(&msg)
			return
		}

		if value, err = p.negCache.Get(key); err != nil {
			Debug("%s didn't hit cache", q)
		} else {
			Debug("%s hit negative cache", q)
			dns.HandleFailed(w, r)
			return
		}
	}

	msg, err := p.lookup(w, r)

	if err != nil {
		Debug("Lookup failed: %s", err)
		dns.HandleFailed(w, r)
		Debug("%s put into negative cache", q)
		p.negCache.Set(key, nil)
		return
	}

	w.WriteMsg(msg)

	if q == nil {
		return
	}

	if len(msg.Answer) > 0 {
		p.cache.Set(key, msg)
		Debug("%s put into cache", q)
	} else {
		p.negCache.Set(key, nil)
		Debug("%s put into negative cache", q)
	}
}

func (p *Proxy) lookup(w dns.ResponseWriter, r *dns.Msg) (*dns.Msg, error) {
	c := p.clientFor(w.RemoteAddr())

	res := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()
		r, rtt, err := c.Exchange(r, nameserver)
		if err != nil {
			Info("error on %s: %s", nameserver, err.Error())
			return
		}

		if r != nil && r.Rcode != dns.RcodeSuccess {
			Debug("failed to get an valid answer on %s", nameserver)
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		} else {
			Debug("resolved on %s (%s) ttl: %d", nameserver, c.Net, rtt)
		}

		select {
		case res <- r:
		default:
		}
	}

	ticker := time.NewTicker(time.Duration(1) * time.Millisecond)
	defer ticker.Stop()

	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range p.nameservers {
		wg.Add(1)
		go L(nameserver)
		select {
		case r := <-res:
			return r, nil
		case <-ticker.C:
			continue
		}
	}

	wg.Wait()
	select {
	case r := <-res:
		return r, nil
	default:
		return nil, fmt.Errorf("resolv failed on %s (%s)", strings.Join(p.nameservers, "; "), c.Net)
	}
}

func (p *Proxy) clientFor(addr net.Addr) *dns.Client {
	if _, ok := addr.(*net.TCPAddr); ok {
		return p.tcp
	}
	return p.udp
}

func keyFor(q *dns.Question) string {
	h := md5.New()
	h.Write([]byte(q.String()))
	x := h.Sum(nil)
	return fmt.Sprintf("%x", x)
}

func NewDNSProxy(addrs []string, timeout, expire time.Duration, cacheSize int) (dns.Handler, error) {
	nameservers, err := parseNameservers(addrs)
	if err != nil {
		return nil, err
	}

	udp := newClient("udp", timeout)
	tcp := newClient("tcp", timeout)

	cache := gcache.New(cacheSize).ARC().Expiration(expire).Build()
	negCache := gcache.New(cacheSize).LRU().Expiration(expire / 2).Build()

	p := &Proxy{nameservers: nameservers,
		udp: udp, tcp: tcp,
		cache: cache, negCache: negCache}
	p.initialize()
	return p, nil
}

func newClient(net string, timeout time.Duration) *dns.Client {
	return &dns.Client{Net: net, ReadTimeout: timeout, WriteTimeout: timeout, SingleInflight: true}
}

func parseNameservers(addrs []string) ([]string, error) {
	nameservers := make([]string, len(addrs))
	for i, ip := range addrs {
		nameservers[i] = strings.TrimSpace(ip)
		if net.ParseIP(nameservers[i]) != nil {
			nameservers[i] = net.JoinHostPort(nameservers[i], defaultDNSPort)
		} else {
			host, port, err := net.SplitHostPort(nameservers[i])
			if err != nil {
				return nil, fmt.Errorf("Invalid address: %s\n", nameservers[i])
			}
			nameservers[i] = net.JoinHostPort(host, port)
		}
	}
	return nameservers, nil
}
