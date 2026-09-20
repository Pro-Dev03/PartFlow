package barcodes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var errExternalNoResult = errors.New("external provider returned no product")

// ExternalProductLookup contains descriptive product metadata only. Prices,
// stock, suppliers, and quantities are intentionally excluded.
type ExternalProductLookup struct {
	Barcode        string   `json:"barcode"`
	Name           string   `json:"name"`
	Brand          string   `json:"brand,omitempty"`
	Category       string   `json:"category,omitempty"`
	Description    string   `json:"description,omitempty"`
	Size           string   `json:"size,omitempty"`
	Specifications []string `json:"specifications,omitempty"`
	Image          string   `json:"image,omitempty"`
	Source         string   `json:"source"`
	Confidence     float64  `json:"confidence"`
	CacheHit       bool     `json:"cache_hit"`
}

type upcItemDBResponse struct {
	Items []struct {
		Title       string   `json:"title"`
		Brand       string   `json:"brand"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Size        string   `json:"size"`
		Image       string   `json:"image"`
		Images      []string `json:"images"`
		Model       string   `json:"model"`
	} `json:"items"`
}

func (s *Service) LookupExternalProduct(ctx context.Context, barcode string) (*ExternalProductLookup, error) {
	code, err := NormalizeBarcode(barcode)
	if err != nil {
		return nil, err
	}

	s.lookupMu.RLock()
	cached, ok := s.lookupCache[code]
	s.lookupMu.RUnlock()
	if ok {
		cached.CacheHit = true
		cached.Source = "cache"
		return &cached, nil
	}

	result, primaryErr := s.provider.Lookup(ctx, code)
	if primaryErr == nil {
		s.lookupMu.Lock()
		s.lookupCache[code] = *result
		s.lookupMu.Unlock()
		return result, nil
	}

	result, webErr := s.webProvider.Lookup(ctx, code)
	if webErr == nil {
		s.lookupMu.Lock()
		s.lookupCache[code] = *result
		s.lookupMu.Unlock()
		return result, nil
	}
	if errors.Is(primaryErr, errExternalNoResult) && (errors.Is(webErr, errExternalNoResult) || errors.Is(webErr, errWebFallbackDisabled)) {
		return nil, fmt.Errorf("product not found in free sources")
	}
	if !errors.Is(webErr, errExternalNoResult) && !errors.Is(webErr, errWebFallbackDisabled) {
		return nil, fmt.Errorf("free barcode provider unavailable: %w", webErr)
	}
	if !errors.Is(primaryErr, errExternalNoResult) {
		return nil, fmt.Errorf("free barcode provider unavailable: %w", primaryErr)
	}
	return nil, fmt.Errorf("product not found in free sources")
}

func externalJSONRequest(ctx context.Context, endpoint string, target interface{}) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "PartFlow/1.0 barcode lookup")
	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("external provider returned status %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
