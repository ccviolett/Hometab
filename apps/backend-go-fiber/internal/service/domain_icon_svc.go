package service

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"hometab/internal/model"
	"hometab/internal/repository"

	"gorm.io/gorm"
)

const fallbackIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><rect width="64" height="64" rx="14" fill="#eef2ff"/><path d="M32 14a18 18 0 1 0 0 36 18 18 0 0 0 0-36Zm12.8 16h-8.1a31 31 0 0 0-1.9-10.1A14.1 14.1 0 0 1 44.8 30ZM32 18.1c1.3 1.9 2.6 6 2.8 11.9h-5.6c.2-5.9 1.5-10 2.8-11.9ZM18 32c0-.7.1-1.4.2-2h8.9a38 38 0 0 0 0 4h-8.9c-.1-.6-.2-1.3-.2-2Zm1.2 4h8.1c.4 4.1 1.1 7.7 1.9 10.1A14.1 14.1 0 0 1 19.2 36Zm8.1-6h-8.1a14.1 14.1 0 0 1 10-10.1A31 31 0 0 0 27.3 30ZM32 45.9c-1.3-1.9-2.6-6-2.8-11.9h5.6c-.2 5.9-1.5 10-2.8 11.9ZM35 32v2h-6v-4h6v2Zm-.2 14.1c.8-2.4 1.5-6 1.9-10.1h8.1a14.1 14.1 0 0 1-10 10.1ZM36.9 34a38 38 0 0 0 0-4h8.9c.1.6.2 1.3.2 2s-.1 1.4-.2 2h-8.9Z" fill="#4f46e5"/></svg>`

type DomainIconSvc struct {
	repo    *repository.DomainIconRepo
	iconDir string
	client  *http.Client
}

type fetchedIcon struct {
	Data        []byte
	ContentType string
	Ext         string
}

func NewDomainIconSvc(repo *repository.DomainIconRepo, iconDir string) *DomainIconSvc {
	return &DomainIconSvc{
		repo:    repo,
		iconDir: iconDir,
		client: &http.Client{
			Timeout: 6 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

func DefaultIconDir(dbPath string) string {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "icons")
}

func (s *DomainIconSvc) List() ([]model.DomainIcon, error) {
	return s.repo.FindAll()
}

func (s *DomainIconSvc) Upload(host string, data []byte) (*model.DomainIcon, error) {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || strings.ContainsAny(host, `/\\`) {
		return nil, errors.New("invalid host")
	}
	if len(data) == 0 {
		return nil, errors.New("icon file is empty")
	}
	if len(data) > 512*1024 {
		return nil, errors.New("icon exceeds 512 KiB limit")
	}
	contentType := detectIconContentType(data)
	if !isSupportedIconType(contentType) {
		return nil, errors.New("unsupported icon content type")
	}
	if contentType == "image/svg+xml" && !safeUploadedSVG(data) {
		return nil, errors.New("unsafe svg content")
	}
	hash := hashBytes(data)
	path, err := s.writeIconFile(host, hash, extensionForContentType(contentType, ""), data)
	if err != nil {
		return nil, err
	}
	existing, findErr := s.repo.FindByHost(host)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		s.removeRelativeFile(path)
		return nil, findErr
	}
	if existing == nil {
		existing = &model.DomainIcon{Host: host}
	}
	oldPath, pendingPath := existing.IconPath, existing.PendingPath
	existing.IconPath = path
	existing.ContentType = contentType
	existing.Hash = hash
	existing.Source = "manual"
	existing.Status = "ready"
	existing.PendingPath = ""
	existing.PendingContentType = ""
	existing.PendingHash = ""
	existing.LastCheckedAt = time.Now()
	existing.ErrorMessage = ""
	if err := s.repo.Save(existing); err != nil {
		s.removeRelativeFile(path)
		return nil, err
	}
	if oldPath != path {
		s.removeRelativeFile(oldPath)
	}
	s.removeRelativeFile(pendingPath)
	return existing, nil
}

func (s *DomainIconSvc) Delete(host string) error {
	host = strings.ToLower(strings.TrimSpace(host))
	item, err := s.repo.FindByHost(host)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(host); err != nil {
		return err
	}
	s.removeRelativeFile(item.IconPath)
	s.removeRelativeFile(item.PendingPath)
	return nil
}

func (s *DomainIconSvc) Retry(host, rawURL string) (*model.DomainIconCheckResponse, error) {
	host = strings.ToLower(strings.TrimSpace(host))
	resolvedHost, _, err := normalizeURL(rawURL)
	if err != nil || resolvedHost != host {
		return nil, errors.New("url host does not match")
	}
	return s.Check(rawURL)
}

func (s *DomainIconSvc) Check(rawURL string) (*model.DomainIconCheckResponse, error) {
	host, normalizedURL, err := normalizeURL(rawURL)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	existing, findErr := s.repo.FindByHost(host)
	if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil, findErr
	}

	icon, fetchErr := s.fetchIcon(normalizedURL)
	if fetchErr != nil {
		if existing != nil {
			existing.LastCheckedAt = now
			existing.ErrorMessage = fetchErr.Error()
			_ = s.repo.Save(existing)
			return &model.DomainIconCheckResponse{
				Host:           host,
				Status:         "failed",
				CurrentIconURL: resolveIconURL(rawURL),
				Error:          fetchErr.Error(),
			}, nil
		}
		item := &model.DomainIcon{
			Host:          host,
			Source:        "fallback",
			Status:        "failed",
			LastCheckedAt: now,
			ErrorMessage:  fetchErr.Error(),
		}
		if err := s.repo.Save(item); err != nil {
			return nil, err
		}
		return &model.DomainIconCheckResponse{Host: host, Status: "failed", Error: fetchErr.Error()}, nil
	}

	hash := hashBytes(icon.Data)
	if existing == nil {
		iconPath, err := s.writeIconFile(host, hash, icon.Ext, icon.Data)
		if err != nil {
			return nil, err
		}
		item := &model.DomainIcon{
			Host:          host,
			IconPath:      iconPath,
			ContentType:   icon.ContentType,
			Hash:          hash,
			Source:        "auto",
			Status:        "ready",
			LastCheckedAt: now,
		}
		if err := s.repo.Save(item); err != nil {
			return nil, err
		}
		return &model.DomainIconCheckResponse{Host: host, Status: "ready", CurrentIconURL: resolveIconURL(rawURL)}, nil
	}

	if existing.Hash == hash && existing.Status == "ready" {
		existing.LastCheckedAt = now
		existing.ErrorMessage = ""
		if err := s.repo.Save(existing); err != nil {
			return nil, err
		}
		return &model.DomainIconCheckResponse{Host: host, Status: "unchanged", CurrentIconURL: resolveIconURL(rawURL)}, nil
	}

	pendingPath, err := s.writeIconFile(host, hash, icon.Ext, icon.Data)
	if err != nil {
		return nil, err
	}
	if existing.IconPath == "" || existing.Status != "ready" {
		existing.IconPath = pendingPath
		existing.ContentType = icon.ContentType
		existing.Hash = hash
		existing.Source = "auto"
		existing.Status = "ready"
		existing.PendingPath = ""
		existing.PendingContentType = ""
		existing.PendingHash = ""
		existing.LastCheckedAt = now
		existing.ErrorMessage = ""
		if err := s.repo.Save(existing); err != nil {
			return nil, err
		}
		return &model.DomainIconCheckResponse{Host: host, Status: "ready", CurrentIconURL: resolveIconURL(rawURL)}, nil
	}

	existing.PendingPath = pendingPath
	existing.PendingContentType = icon.ContentType
	existing.PendingHash = hash
	existing.LastCheckedAt = now
	existing.ErrorMessage = ""
	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}
	return &model.DomainIconCheckResponse{
		Host:           host,
		Status:         "conflict",
		CurrentIconURL: resolveIconURL(rawURL),
		PendingIconURL: pendingIconURL(host),
	}, nil
}

func (s *DomainIconSvc) Choose(host, choice string) (*model.DomainIcon, error) {
	host = strings.ToLower(strings.TrimSpace(host))
	item, err := s.repo.FindByHost(host)
	if err != nil {
		return nil, err
	}
	switch choice {
	case "current":
		s.removeRelativeFile(item.PendingPath)
		item.PendingPath = ""
		item.PendingContentType = ""
		item.PendingHash = ""
	case "new":
		if item.PendingPath == "" || item.PendingHash == "" {
			return nil, errors.New("no pending icon")
		}
		oldPath := item.IconPath
		item.IconPath = item.PendingPath
		item.ContentType = item.PendingContentType
		item.Hash = item.PendingHash
		item.Source = "user_confirmed"
		item.Status = "ready"
		item.PendingPath = ""
		item.PendingContentType = ""
		item.PendingHash = ""
		s.removeRelativeFile(oldPath)
	default:
		return nil, errors.New("choice must be current or new")
	}
	item.ErrorMessage = ""
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *DomainIconSvc) Resolve(rawURL string) ([]byte, string, error) {
	host, _, err := normalizeURL(rawURL)
	if err != nil {
		return []byte(fallbackIconSVG), "image/svg+xml", nil
	}
	item, err := s.repo.FindByHost(host)
	if err != nil || item.IconPath == "" || item.Status != "ready" {
		return []byte(fallbackIconSVG), "image/svg+xml", nil
	}
	data, err := os.ReadFile(s.absoluteIconPath(item.IconPath))
	if err != nil {
		return []byte(fallbackIconSVG), "image/svg+xml", nil
	}
	contentType := item.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return data, contentType, nil
}

func (s *DomainIconSvc) Pending(host string) ([]byte, string, error) {
	item, err := s.repo.FindByHost(strings.ToLower(strings.TrimSpace(host)))
	if err != nil || item.PendingPath == "" {
		return []byte(fallbackIconSVG), "image/svg+xml", nil
	}
	data, err := os.ReadFile(s.absoluteIconPath(item.PendingPath))
	if err != nil {
		return []byte(fallbackIconSVG), "image/svg+xml", nil
	}
	contentType := item.PendingContentType
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	return data, contentType, nil
}
