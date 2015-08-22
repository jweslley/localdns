package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var (
	programVersion = "dev"
	ip4loopback    = net.IPv4(127, 0, 0, 1)

	// flags
	tld     = flag.String("tld", "dev", "Top-level domain to resolve to localhost")
	ttl     = flag.Int("ttl", 600, "DNS's TTL (Time to live)")
	port    = flag.Int("port", 5353, "DNS's port")
	version = flag.Bool("v", false, "print version information and exit")
)

func localdns(w dns.ResponseWriter, r *dns.Msg) {
	if r.Opcode != dns.OpcodeQuery || len(r.Question) == 0 {
		dns.HandleFailed(w, r)
		return
	}

	q := r.Question[0]
	if !(q.Qclass == dns.ClassINET && (q.Qtype == dns.TypeA || q.Qtype == dns.TypeAAAA)) {
		dns.HandleFailed(w, r)
		return
	}

	domain := dns.Fqdn(q.Name)
	if !strings.HasSuffix(domain, *tld) {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

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

	*tld = dns.Fqdn(*tld)
	if (*tld)[0] != '.' {
		*tld = fmt.Sprintf(".%s", *tld)
	}

	dns.HandleFunc(".", localdns)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting localdns at %s\n", addr)
	log.Printf("localdns will respond 'localhost' to queries in %s domain", *tld)
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
