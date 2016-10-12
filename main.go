package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var (
	programVersion = "dev"
	ip4loopback    = net.IPv4(127, 0, 0, 1)

	// flags
	tld  = flag.String("tld", "dev", "comma-separated list of top-level domains to resolve to localhost. Example: dev,test")
	ttl  = flag.Int("ttl", 600, "DNS's TTL (Time to live)")
	port = flag.Int("port", 5353, "DNS's port")

	proxy     = flag.String("proxy", "", "enable proxy mode. comma-separated list of DNS servers to send queries for. Example: 8.8.8.8,8.8.4.4")
	timeout   = flag.Duration("timeout", time.Duration(2)*time.Second, "when acting as proxy, timeout for dial, write and read. Example: 3s")
	expire    = flag.Duration("expire", time.Duration(10)*time.Minute, "when acting as proxy, cache expiration time. Example: 5m")
	cacheSize = flag.Int("cache", 65536, "when acting as proxy, the cache size.")

	debug   = flag.Bool("debug", false, "enable verbose logging")
	version = flag.Bool("v", false, "print version information and exit")
)

func Debug(format string, v ...interface{}) {
	if *debug {
		log.Printf(format, v...)
	}
}

func Info(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func localdns(w dns.ResponseWriter, r *dns.Msg) {
	q, ok := questionFor(r)
	if !ok {
		dns.HandleFailed(w, r)
		return
	}

	Debug("%s lookup %s\n", w.RemoteAddr(), q)

	m := new(dns.Msg)
	m.SetReply(r)

	rr := dns.RR_Header{Name: q.Name, Class: dns.ClassINET, Rrtype: q.Qtype, Ttl: uint32(*ttl)}
	if q.Qtype == dns.TypeA {
		m.Answer = append(m.Answer, &dns.A{Hdr: rr, A: ip4loopback})
	} else {
		m.Answer = append(m.Answer, &dns.AAAA{Hdr: rr, AAAA: net.IPv6loopback})
	}

	w.WriteMsg(m)
}

func questionFor(r *dns.Msg) (*dns.Question, bool) {
	if r.Opcode != dns.OpcodeQuery || len(r.Question) == 0 {
		return nil, false
	}

	q := r.Question[0]
	if !(q.Qclass == dns.ClassINET && (q.Qtype == dns.TypeA || q.Qtype == dns.TypeAAAA)) {
		return nil, false
	}

	return &q, true
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: localdns [OPTION]...\n")
	fmt.Fprintf(os.Stderr, "A DNS for local development.\n\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Printf("localdns %s\n", programVersion)
		return
	}

	*tld = strings.TrimSpace(*tld)
	if *tld == "" {
		log.Fatalln("localdns: tld required")
	}

	tlds := strings.Split(*tld, ",")
	for _, tld := range tlds {
		tld = dns.Fqdn(tld)
		Info("localdns will respond 'localhost' to queries in %s", tld)
		dns.HandleFunc(tld, localdns)
	}

	*proxy = strings.TrimSpace(*proxy)
	if len(*proxy) > 0 {
		proxy, err := NewDNSProxy(strings.Split(*proxy, ","), *timeout, *expire, *cacheSize)
		if err != nil {
			log.Fatalf("localdns: %v\n", err)
		}
		Info("Proxy mode enabled.")
		dns.Handle(".", proxy)
	}

	addr := fmt.Sprintf(":%d", *port)
	Info("Starting localdns at %s\n", addr)
	go func() {
		err := dns.ListenAndServe(addr, "udp", nil)
		if err != nil {
			log.Fatalf("localdns: Failed to set udp listener: %s\n", err)
		}
	}()

	err := dns.ListenAndServe(fmt.Sprintf(":%d", *port), "tcp", nil)
	if err != nil {
		log.Fatalf("localdns: Failed to set tcp listener: %s\n", err)
	}
}
