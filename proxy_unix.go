// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"
)

func (p *Proxy) initialize() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for sig := range c {
			go func(sig os.Signal) {
				switch sig {
				case syscall.SIGUSR1:
					p.flushAll()
				case syscall.SIGUSR2:
					p.printStats()
				}
			}(sig)
		}
	}()
}

func (p *Proxy) flushAll() {
	p.cache.Purge()
	p.negCache.Purge()
	Info("Cache was cleared")
}

func (p *Proxy) printStats() {
	Info("Cache stats: HitCount=%d, MissCount=%d, LookupCount=%d, HitRate=%f",
		p.cache.HitCount(),
		p.cache.MissCount(),
		p.cache.LookupCount(),
		p.cache.HitRate())
}
