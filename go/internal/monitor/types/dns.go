package types

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type DNSChecker struct{}

func (c *DNSChecker) Type() string { return "dns" }

func (c *DNSChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	resolver := &net.Resolver{
		PreferGo: true,
	}
	if cfg.DNS.ResolveType != "" && cfg.Hostname != "" {
		server := cfg.Hostname
		if cfg.Port > 0 {
			server = fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)
		} else {
			server = server + ":53"
		}
		resolver.Dial = func(ctx context.Context, network, _ string) (net.Conn, error) {
			d := net.Dialer{Timeout: timeout}
			return d.DialContext(ctx, "udp", server)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	hostname := cfg.URL
	if hostname == "" {
		hostname = cfg.Hostname
	}

	start := time.Now()
	var records []string
	var lookupErr error

	switch strings.ToUpper(cfg.DNS.ResolveType) {
	case "A", "AAAA", "":
		var ips []net.IP
		ips, lookupErr = resolver.LookupIP(ctx, resolveNetwork(cfg.DNS.ResolveType), hostname)
		for _, ip := range ips {
			records = append(records, ip.String())
		}
	case "MX":
		var mxs []*net.MX
		mxs, lookupErr = resolver.LookupMX(ctx, hostname)
		for _, mx := range mxs {
			records = append(records, fmt.Sprintf("%s %d", mx.Host, mx.Pref))
		}
	case "TXT":
		records, lookupErr = resolver.LookupTXT(ctx, hostname)
	case "CNAME":
		var cname string
		cname, lookupErr = resolver.LookupCNAME(ctx, hostname)
		if cname != "" {
			records = append(records, cname)
		}
	case "NS":
		var nss []*net.NS
		nss, lookupErr = resolver.LookupNS(ctx, hostname)
		for _, ns := range nss {
			records = append(records, ns.Host)
		}
	case "SRV":
		_, srvs, srvErr := resolver.LookupSRV(ctx, "", "", hostname)
		lookupErr = srvErr
		for _, srv := range srvs {
			records = append(records, fmt.Sprintf("%s:%d", srv.Target, srv.Port))
		}
	default:
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("unsupported DNS resolve type: %s", cfg.DNS.ResolveType),
		}, nil
	}

	ping := time.Since(start).Milliseconds()

	if lookupErr != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: fmt.Sprintf("DNS lookup failed: %v", lookupErr),
		}, nil
	}

	if len(records) == 0 {
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: "no records found",
		}, nil
	}

	return monitor.CheckResult{
		Status:  status.Up,
		Ping:    ping,
		Message: strings.Join(records, ", "),
	}, nil
}

func resolveNetwork(resolveType string) string {
	switch strings.ToUpper(resolveType) {
	case "AAAA":
		return "ip6"
	case "A":
		return "ip4"
	default:
		return "ip"
	}
}
