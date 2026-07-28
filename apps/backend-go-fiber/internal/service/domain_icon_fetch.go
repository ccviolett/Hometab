package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func (s *DomainIconSvc) fetchIcon(rawURL string) (*fetchedIcon, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	candidates := []string{
		u.Scheme + "://" + u.Host + "/favicon.ico",
		u.Scheme + "://" + u.Host + "/favicon.png",
		u.Scheme + "://" + u.Host + "/apple-touch-icon.png",
	}
	for _, candidate := range candidates {
		if icon, err := s.downloadIcon(candidate); err == nil {
			return icon, nil
		}
	}
	htmlCandidates, _ := s.iconLinksFromHTML(u.Scheme + "://" + u.Host + "/")
	for _, candidate := range htmlCandidates {
		if icon, err := s.downloadIcon(candidate); err == nil {
			return icon, nil
		}
	}
	return nil, errors.New("icon not found")
}

func (s *DomainIconSvc) iconLinksFromHTML(pageURL string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "HomeIconResolver/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("html status %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, 512*1024)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	base, _ := url.Parse(pageURL)
	matches := iconLinkRe.FindAllSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		tag := string(match[0])
		if !relIconRe.MatchString(tag) {
			continue
		}
		hrefMatch := hrefRe.FindStringSubmatch(tag)
		if len(hrefMatch) < 2 {
			continue
		}
		href := html.UnescapeString(hrefMatch[1])
		parsed, err := url.Parse(href)
		if err != nil {
			continue
		}
		out = append(out, base.ResolveReference(parsed).String())
	}
	return out, nil
}

var (
	iconLinkRe = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	relIconRe  = regexp.MustCompile(`(?i)\brel\s*=\s*["'][^"']*(icon|shortcut icon|apple-touch-icon)[^"']*["']`)
	hrefRe     = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)
)

func (s *DomainIconSvc) downloadIcon(iconURL string) (*fetchedIcon, error) {
	req, err := http.NewRequest(http.MethodGet, iconURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "HomeIconResolver/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("icon status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("empty icon")
	}
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}
	if !isSupportedIconType(contentType) {
		return nil, fmt.Errorf("unsupported icon content type: %s", contentType)
	}
	ext := extensionForContentType(contentType, iconURL)
	return &fetchedIcon{Data: data, ContentType: contentType, Ext: ext}, nil
}

func normalizeURL(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("url is required")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", "", errors.New("invalid url")
	}
	host := strings.ToLower(u.Host)
	u.Host = host
	return host, u.String(), nil
}

func HostForURL(raw string) (string, error) {
	host, _, err := normalizeURL(raw)
	return host, err
}

func resolveIconURL(rawURL string) string {
	return "/api/link-icons/resolve?url=" + url.QueryEscape(rawURL)
}

func pendingIconURL(host string) string {
	return "/api/link-icons/pending/" + url.PathEscape(host)
}

func (s *DomainIconSvc) writeIconFile(host, hash, ext string, data []byte) (string, error) {
	if err := os.MkdirAll(s.iconDir, 0755); err != nil {
		return "", err
	}
	filename := sanitizeHost(host) + "-" + hash[:12] + ext
	if err := os.WriteFile(filepath.Join(s.iconDir, filename), data, 0644); err != nil {
		return "", err
	}
	return filename, nil
}

func (s *DomainIconSvc) absoluteIconPath(relative string) string {
	return filepath.Join(s.iconDir, filepath.Base(relative))
}

func (s *DomainIconSvc) removeRelativeFile(relative string) {
	if relative == "" {
		return
	}
	_ = os.Remove(s.absoluteIconPath(relative))
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func sanitizeHost(host string) string {
	replacer := strings.NewReplacer(":", "_", ".", "_", "[", "", "]", "")
	return replacer.Replace(strings.ToLower(host))
}

func detectIconContentType(data []byte) string {
	prefix := strings.ToLower(strings.TrimSpace(string(data[:min(len(data), 512)])))
	if strings.Contains(prefix, "<svg") {
		return "image/svg+xml"
	}
	return strings.Split(http.DetectContentType(data), ";")[0]
}

func safeUploadedSVG(data []byte) bool {
	lower := strings.ToLower(string(data))
	for _, denied := range []string{"<script", "javascript:", "onload=", "onerror=", "<foreignobject", `href="http:`, `href='http:`, `href="https:`, `href='https:`, "xlink:href"} {
		if strings.Contains(lower, denied) {
			return false
		}
	}
	return true
}

func isSupportedIconType(contentType string) bool {
	switch contentType {
	case "image/x-icon", "image/vnd.microsoft.icon", "image/png", "image/jpeg", "image/svg+xml", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func extensionForContentType(contentType, rawURL string) string {
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		if exts[0] == ".jpe" {
			return ".jpg"
		}
		return exts[0]
	}
	ext := strings.ToLower(path.Ext(rawURL))
	switch ext {
	case ".ico", ".png", ".svg", ".jpg", ".jpeg", ".gif", ".webp":
		return ext
	default:
		return ".ico"
	}
}
