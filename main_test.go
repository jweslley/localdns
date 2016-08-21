package main

import (
	"net"
	"sync"
	"testing"

	"github.com/miekg/dns"
)

func TestLocalDNS(t *testing.T) {
	dns.HandleFunc("dev.", localdns)

	s, addrstr, err := RunLocalUDPServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Unable to run test server: %v", err)
	}
	defer s.Shutdown()

	c := new(dns.Client)
	lookupCheckIPv4 := func(question string, answer net.IP) {
		m := new(dns.Msg)
		m.SetQuestion(question, dns.TypeA)
		r, _, err := c.Exchange(m, addrstr)
		if err != nil || len(r.Answer) == 0 {
			t.Fatalf("failed to exchange %s: %v", question, err)
		}
		a := r.Answer[0].(*dns.A)
		if a.A.String() != answer.String() {
			t.Errorf("Wrong answer: got %d; expected %d", a.A, answer)
		}
	}

	lookupCheckIPv6 := func(question string, answer net.IP) {
		m := new(dns.Msg)
		m.SetQuestion(question, dns.TypeAAAA)
		r, _, err := c.Exchange(m, addrstr)
		if err != nil || len(r.Answer) == 0 {
			t.Fatalf("failed to exchange %s: %v", question, err)
		}
		a := r.Answer[0].(*dns.AAAA)
		if a.AAAA.String() != answer.String() {
			t.Errorf("Wrong answer: got %d; expected %d", a.AAAA, answer)
		}
	}

	questions := []string{"myapp.dev.", "subdomain.myapp.dev.", "pt.subdomain.myapp.dev."}

	for _, q := range questions {
		lookupCheckIPv4(q, ip4loopback)
		lookupCheckIPv6(q, net.IPv6loopback)
	}

	failedLookupCheck := func(question string, qtype uint16) {
		m := new(dns.Msg)
		m.SetQuestion(question, qtype)
		r, _, err := c.Exchange(m, addrstr)
		if err != nil {
			t.Fatalf("failed to exchange %s: %v", question, err)
		}
		if r.Rcode != dns.RcodeServerFailure {
			t.Errorf("Wrong rocde: got %d; expected %d", r.Rcode, dns.RcodeNameError)
		}
	}

	questions = []string{"myapp.local.", "subdomain.myapp.lol.", "pt.subdomain.myapp.com."}

	for _, q := range questions {
		failedLookupCheck(q, dns.TypeA)
		failedLookupCheck(q, dns.TypeAAAA)
	}
}

func RunLocalUDPServer(laddr string) (*dns.Server, string, error) {
	pc, err := net.ListenPacket("udp", laddr)
	if err != nil {
		return nil, "", err
	}
	server := &dns.Server{PacketConn: pc}

	waitLock := sync.Mutex{}
	waitLock.Lock()
	server.NotifyStartedFunc = waitLock.Unlock

	go func() {
		server.ActivateAndServe()
		pc.Close()
	}()

	waitLock.Lock()
	return server, pc.LocalAddr().String(), nil
}
