package browser

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ErrBlockedURL is returned when a URL targets an internal or private network address.
var ErrBlockedURL = errors.New("URL targets a blocked internal/private network address")

// ErrEvalBlockedP2P is returned when eval is attempted from a P2P peer context.
var ErrEvalBlockedP2P = errors.New("eval action is not permitted for remote peer requests")

// privateNetworks defines CIDR ranges considered internal/private.
var privateNetworks = []net.IPNet{
	// 10.0.0.0/8
	{IP: net.IP{10, 0, 0, 0}, Mask: net.CIDRMask(8, 32)},
	// 172.16.0.0/12
	{IP: net.IP{172, 16, 0, 0}, Mask: net.CIDRMask(12, 32)},
	// 192.168.0.0/16
	{IP: net.IP{192, 168, 0, 0}, Mask: net.CIDRMask(16, 32)},
	// 169.254.0.0/16 (link-local)
	{IP: net.IP{169, 254, 0, 0}, Mask: net.CIDRMask(16, 32)},
	// 127.0.0.0/8 (loopback)
	{IP: net.IP{127, 0, 0, 0}, Mask: net.CIDRMask(8, 32)},
}

// ValidateURLForP2P checks that a URL is safe for navigation in a P2P context.
// It blocks file:// schemes and URLs that resolve to internal/private network addresses.
func ValidateURLForP2P(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	// Block file:// scheme.
	if strings.EqualFold(parsed.Scheme, "file") {
		return fmt.Errorf("%w: file:// scheme is not allowed", ErrBlockedURL)
	}

	// Extract hostname (strip port if present).
	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("%w: empty hostname", ErrBlockedURL)
	}

	// Block localhost by name.
	lower := strings.ToLower(hostname)
	if lower == "localhost" {
		return fmt.Errorf("%w: localhost is not allowed", ErrBlockedURL)
	}

	// Block IPv6 loopback [::1].
	if lower == "::1" {
		return fmt.Errorf("%w: IPv6 loopback is not allowed", ErrBlockedURL)
	}

	// Parse as IP and check against private ranges.
	ip := net.ParseIP(hostname)
	if ip != nil {
		if err := checkIPPrivate(ip, hostname); err != nil {
			return err
		}
	} else {
		// Hostname is not an IP literal — resolve via DNS and check all results.
		ips, err := net.LookupIP(hostname)
		if err == nil {
			for _, resolved := range ips {
				if err := checkIPPrivate(resolved, hostname); err != nil {
					return err
				}
			}
		}
		// If DNS lookup fails, allow the request — the browser will fail on its own.
	}

	return nil
}

// checkIPPrivate returns an error if ip falls within a private/loopback range.
func checkIPPrivate(ip net.IP, label string) error {
	if ip.IsLoopback() {
		return fmt.Errorf("%w: loopback address is not allowed", ErrBlockedURL)
	}
	for _, cidr := range privateNetworks {
		if cidr.Contains(ip) {
			return fmt.Errorf("%w: %s resolves to a private network address", ErrBlockedURL, label)
		}
	}
	return nil
}
