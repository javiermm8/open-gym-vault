package api

import (
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"sync"
)

var trusted []netip.Prefix

var loadTrustedProxies = sync.OnceValue(func() []netip.Prefix {
	trusted = parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	return trusted
})

func parseTrustedProxies(raw string) []netip.Prefix {
	if raw == "" {
		return nil
	}

	var prefixes []netip.Prefix
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if p, err := netip.ParsePrefix(part); err == nil {
			prefixes = append(prefixes, p)
			continue
		}
		if addr, err := netip.ParseAddr(part); err == nil {
			addr = addr.Unmap()
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		log.Printf("trusted proxies: ignoring invalid entry %q", part)
	}
	if len(prefixes) == 0 {
		log.Printf("trusted proxies: no valid entries, X-Forwarded-For will be ignored")
	}
	return prefixes
}

func isTrustedProxy(addr netip.Addr) bool {
	if trusted == nil {
		loadTrustedProxies()
	}
	addr = addr.Unmap()
	for _, p := range trusted {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	remote, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}
	if !isTrustedProxy(remote) {
		return remote.String()
	}

	entries := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	for i := len(entries) - 1; i >= 0; i-- {
		addr, err := netip.ParseAddr(strings.TrimSpace(entries[i]))
		if err != nil {
			continue
		}
		if !isTrustedProxy(addr) {
			return addr.String()
		}
	}
	return remote.String()
}
