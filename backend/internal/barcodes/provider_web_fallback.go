package barcodes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var errWebFallbackDisabled = errors.New("web fallback is not configured")

const maxPublicPageBytes = 2 << 20

type searxngSearchProvider struct {
	baseURL string
	client  *http.Client
}

type searxngResponse struct {
	Results []struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		URL     string `json:"url"`
	} `json:"results"`
}

type publicProductPage struct {
	Name        string
	Brand       string
	Category    string
	Description string
	Image       string
}

func NewWebFallbackProvider() BarcodeProvider {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("SEARXNG_URL")), "/")
	if baseURL == "" {
		baseURL = "http://searxng:8080"
	}
	return &searxngSearchProvider{baseURL: baseURL, client: &http.Client{Timeout: 8 * time.Second}}
}

func (p *searxngSearchProvider) Lookup(ctx context.Context, barcode string) (*ProductLookup, error) {
	if p.baseURL == "" {
		return nil, errWebFallbackDisabled
	}

	searchURL := p.baseURL + "/search?q=" + url.QueryEscape(barcode) + "&format=json&categories=general"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "PartFlow/1.0 barcode lookup")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("SearXNG returned status %d", response.StatusCode)
	}

	var payload searxngResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	for _, result := range payload.Results {
		if !containsExactBarcode(result.Title, barcode) && !containsExactBarcode(result.Content, barcode) && !containsExactBarcode(result.URL, barcode) {
			continue
		}
		page, err := p.readPublicProductPage(ctx, result.URL, barcode)
		if err != nil {
			continue
		}
		return &ProductLookup{
			Barcode:     barcode,
			Name:        page.Name,
			Brand:       page.Brand,
			Category:    page.Category,
			Description: page.Description,
			Image:       page.Image,
			Source:      "searxng-web",
			Confidence:  0.6,
		}, nil
	}
	return nil, errExternalNoResult
}

func (p *searxngSearchProvider) readPublicProductPage(ctx context.Context, rawURL, barcode string) (*publicProductPage, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return nil, fmt.Errorf("public result URL is not an HTTPS page")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml")
	request.Header.Set("User-Agent", "PartFlow/1.0 barcode lookup")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("public page returned status %d", response.StatusCode)
	}
	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if !strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("public result is not HTML")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxPublicPageBytes))
	if err != nil {
		return nil, err
	}
	text := cleanWebText(string(body))
	if !containsExactBarcode(text, barcode) {
		return nil, errExternalNoResult
	}

	page := &publicProductPage{
		Name: firstNonEmpty(
			extractMeta(string(body), "og:title"),
			extractJSONLDName(string(body)),
			extractTitle(string(body), "h1"),
		),
		Brand:       extractJSONLDField(string(body), "brand"),
		Category:    extractJSONLDField(string(body), "category"),
		Description: firstNonEmpty(extractMeta(string(body), "og:description"), extractMeta(string(body), "description")),
		Image:       extractMeta(string(body), "og:image"),
	}
	if page.Name == "" {
		return nil, errExternalNoResult
	}
	return page, nil
}

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)
var metaPattern = regexp.MustCompile(`(?is)<meta[^>]+(?:property|name)=["']([^"']+)["'][^>]+content=["']([^"']*)["'][^>]*>`)
var titlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
var h1Pattern = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
var jsonLDNamePattern = regexp.MustCompile(`(?is)"name"\s*:\s*"([^"]+)"`)
var jsonLDFieldPattern = regexp.MustCompile(`(?is)"%s"\s*:\s*(?:\{\s*"name"\s*:\s*)?"([^"]+)"`)

func cleanWebText(value string) string {
	return strings.TrimSpace(html.UnescapeString(htmlTagPattern.ReplaceAllString(value, " ")))
}

func extractMeta(body, key string) string {
	for _, match := range metaPattern.FindAllStringSubmatch(body, -1) {
		if strings.EqualFold(strings.TrimSpace(match[1]), key) {
			return cleanWebText(match[2])
		}
	}
	return ""
}

func extractTitle(body, tag string) string {
	pattern := titlePattern
	if tag == "h1" {
		pattern = h1Pattern
	}
	match := pattern.FindStringSubmatch(body)
	if len(match) == 2 {
		return cleanWebText(match[1])
	}
	return ""
}

func extractJSONLDName(body string) string {
	match := jsonLDNamePattern.FindStringSubmatch(body)
	if len(match) == 2 {
		return cleanWebText(match[1])
	}
	return ""
}

func extractJSONLDField(body, field string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(jsonLDFieldPattern.String(), regexp.QuoteMeta(field)))
	match := pattern.FindStringSubmatch(body)
	if len(match) == 2 {
		return cleanWebText(match[1])
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func containsExactBarcode(value, barcode string) bool {
	for _, token := range strings.FieldsFunc(value, func(r rune) bool { return r < '0' || r > '9' }) {
		if token == barcode {
			return true
		}
	}
	return false
}
