package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	externalRequestTimeout      = 10 * time.Second
	externalRequestMaxRedirects = 3
	externalRequestConcurrency  = 4
)

var deniedExternalPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

var metadataHosts = map[string]bool{
	"metadata.google.internal": true,
	"metadata.azure.internal":  true,
	"instance-data":            true,
}

type externalRequestHTTP struct {
	client   *http.Client
	resolver *net.Resolver
	sem      chan struct{}
}

func newExternalRequestHTTP() *externalRequestHTTP {
	h := &externalRequestHTTP{
		resolver: net.DefaultResolver,
		sem:      make(chan struct{}, externalRequestConcurrency),
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           h.dialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          16,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 8 * time.Second,
	}
	h.client = &http.Client{
		Transport: transport,
		Timeout:   externalRequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= externalRequestMaxRedirects {
				return errors.New("redirect limit exceeded")
			}
			return validateExternalTarget(req.Context(), h.resolver, req.URL)
		},
	}
	return h
}

func (h *externalRequestHTTP) do(req *http.Request) (*http.Response, error) {
	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	default:
		return nil, errors.New("concurrency_limited")
	}
	if err := validateExternalTarget(req.Context(), h.resolver, req.URL); err != nil {
		return nil, err
	}
	return h.client.Do(req)
}

func (h *externalRequestHTTP) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, errors.New("target_denied")
	}
	addresses, err := resolveExternalAddresses(ctx, h.resolver, host)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	var lastErr error
	for _, addr := range addresses {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(addr.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("target_denied")
	}
	return nil, lastErr
}

func validateExternalTarget(ctx context.Context, resolver *net.Resolver, target *url.URL) error {
	if target == nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" {
		return errors.New("invalid_url")
	}
	if target.User != nil {
		return errors.New("target_denied")
	}
	host := strings.ToLower(strings.TrimSuffix(target.Hostname(), "."))
	if metadataHosts[host] {
		return errors.New("target_denied")
	}
	_, err := resolveExternalAddresses(ctx, resolver, host)
	return err
}

func resolveExternalAddresses(ctx context.Context, resolver *net.Resolver, host string) ([]netip.Addr, error) {
	if literal, err := netip.ParseAddr(host); err == nil {
		literal = literal.Unmap()
		if externalAddressDenied(literal) {
			return nil, errors.New("target_denied")
		}
		return []netip.Addr{literal}, nil
	}
	ips, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, errors.New("target_denied")
	}
	result := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		ip = ip.Unmap()
		if externalAddressDenied(ip) {
			return nil, errors.New("target_denied")
		}
		result = append(result, ip)
	}
	return result, nil
}

func externalAddressDenied(addr netip.Addr) bool {
	if !addr.IsValid() || addr.IsUnspecified() || addr.IsMulticast() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() {
		return true
	}
	for _, prefix := range deniedExternalPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func IsLoopbackHost(host string) bool {
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return true
	}
	addr, err := netip.ParseAddr(host)
	return err == nil && addr.Unmap().IsLoopback()
}

func externalRequestError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	for _, code := range []string{"invalid_url", "target_denied", "concurrency_limited"} {
		if strings.Contains(message, code) {
			return code
		}
	}
	if strings.Contains(message, "redirect limit exceeded") {
		return "redirect_denied"
	}
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(message), "timeout") {
		return "timeout"
	}
	return fmt.Sprintf("upstream_error")
}
