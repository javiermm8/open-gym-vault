package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies string
		remoteAddr     string
		xff            string
		want           string
	}{
		{
			name:       "direct connection",
			remoteAddr: "203.0.113.7:54321",
			want:       "203.0.113.7",
		},
		{
			name:       "untrusted peer, spoofed xff is ignored",
			remoteAddr: "203.0.113.7:54321",
			xff:        "1.2.3.4",
			want:       "203.0.113.7",
		},
		{
			name:           "trusted proxy, single xff entry",
			trustedProxies: "203.0.113.7",
			remoteAddr:     "203.0.113.7:54321",
			xff:            "198.51.100.9",
			want:           "198.51.100.9",
		},
		{
			name:           "trusted chain, rightmost untrusted wins",
			trustedProxies: "203.0.113.7,198.51.100.0/24",
			remoteAddr:     "203.0.113.7:54321",
			xff:            "1.2.3.4, 198.51.100.9, 198.51.100.10",
			want:           "1.2.3.4",
		},
		{
			name:           "trusted proxy, client spoofed entry is not leftmost-untrusted",
			trustedProxies: "203.0.113.7",
			remoteAddr:     "203.0.113.7:54321",
			xff:            "6.6.6.6, 198.51.100.9",
			want:           "198.51.100.9",
		},
		{
			name:           "trusted proxy, all entries trusted, falls back to peer",
			trustedProxies: "203.0.113.7,198.51.100.0/24",
			remoteAddr:     "203.0.113.7:54321",
			xff:            "198.51.100.9, 198.51.100.10",
			want:           "203.0.113.7",
		},
		{
			name:           "trusted proxy, no xff header",
			trustedProxies: "203.0.113.7",
			remoteAddr:     "203.0.113.7:54321",
			want:           "203.0.113.7",
		},
		{
			name:           "trusted proxy, garbage xff ignored",
			trustedProxies: "203.0.113.7",
			remoteAddr:     "203.0.113.7:54321",
			xff:            "not-an-ip",
			want:           "203.0.113.7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// isTrustedProxy caches the parsed list; inject a fresh one
			// per case instead of relying on the process env.
			trusted = parseTrustedProxies(tt.trustedProxies)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}

			if got := clientIP(r); got != tt.want {
				t.Errorf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
