package barcodes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubBarcodeProvider struct {
	result   *ProductLookup
	err      error
	calls    int
	lastCode string
}

func (p *stubBarcodeProvider) Lookup(_ context.Context, barcode string) (*ProductLookup, error) {
	p.calls++
	p.lastCode = barcode
	return p.result, p.err
}

func TestWebFallbackRequiresExactBarcode(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"title":"PaperOne Copier A4 80g 8993242596979","content":"Product page","url":"https://` + r.Host + `/product"}]}`))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><meta property="og:title" content="PaperOne Copier A4 80g"><meta property="og:description" content="8993242596979 copier paper"><script type="application/ld+json">{"@type":"Product","name":"PaperOne Copier A4 80g","brand":{"name":"PaperOne"}}</script></head><body><h1>PaperOne Copier A4 80g</h1><p>8993242596979</p></body></html>`))
	}))
	defer server.Close()

	provider := &searxngSearchProvider{baseURL: server.URL, client: server.Client()}
	result, err := provider.Lookup(context.Background(), "8993242596979")
	if err != nil || result == nil {
		t.Fatalf("expected exact result, got %v", err)
	}
	if result.Source != "searxng-web" || result.Barcode != "8993242596979" || result.Name != "PaperOne Copier A4 80g" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestWebFallbackRejectsSimilarBarcode(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"PaperOne Copier A4 80g 8993242596978","content":"Similar product","url":"https://` + r.Host + `/product"}]}`))
	}))
	defer server.Close()

	provider := &searxngSearchProvider{baseURL: server.URL, client: server.Client()}
	_, err := provider.Lookup(context.Background(), "8993242596979")
	if !errors.Is(err, errExternalNoResult) {
		t.Fatalf("expected exact-match no result, got %v", err)
	}
}

func TestWebFallbackFindsTascoIcedCoffeeByExactBarcode(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"title":"Tasco Iced Coffee 8854419001507","content":"Public product page","url":"https://` + r.Host + `/tasco"}]}`))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><meta property="og:title" content="Tasco Iced Coffee"><meta property="og:description" content="8854419001507 iced coffee"><script type="application/ld+json">{"@type":"Product","name":"Tasco Iced Coffee","brand":{"name":"Tasco"}}</script></head><body><h1>Tasco Iced Coffee</h1><p>Barcode: 8854419001507</p></body></html>`))
	}))
	defer server.Close()

	provider := &searxngSearchProvider{baseURL: server.URL, client: server.Client()}
	result, err := provider.Lookup(context.Background(), "8854419001507")
	if err != nil || result == nil || result.Name != "Tasco Iced Coffee" || result.Barcode != "8854419001507" {
		t.Fatalf("expected Tasco exact result, got result=%+v err=%v", result, err)
	}
	if result.Source != "searxng-web" {
		t.Fatalf("expected searxng-web source, got %q", result.Source)
	}
}

func TestServiceFallsBackOnlyAfterProviderNoResult(t *testing.T) {
	primary := &stubBarcodeProvider{err: errExternalNoResult}
	web := &stubBarcodeProvider{result: &ProductLookup{Barcode: "0123456789012", Name: "Fallback Product", Source: "web-fallback"}}
	service := &Service{provider: primary, webProvider: web, lookupCache: make(map[string]ProductLookup)}

	result, err := service.LookupExternalProduct(context.Background(), " 0123-456789012 ")
	if err != nil || result == nil || result.Source != "web-fallback" {
		t.Fatalf("expected web fallback result, got result=%+v err=%v", result, err)
	}
	if primary.calls != 1 || primary.lastCode != "0123456789012" || web.calls != 1 || web.lastCode != "0123456789012" {
		t.Fatalf("providers did not receive normalized barcode: primary=%+v web=%+v", primary, web)
	}

	_, err = service.LookupExternalProduct(context.Background(), "0123456789012")
	if err != nil || web.calls != 1 {
		t.Fatalf("expected cached result without provider call, calls=%d err=%v", web.calls, err)
	}
}

func TestServiceStopsAfterPrimaryProviderSuccess(t *testing.T) {
	primary := &stubBarcodeProvider{result: &ProductLookup{Barcode: "036000291452", Name: "Primary Product", Source: "upcitemdb"}}
	web := &stubBarcodeProvider{result: &ProductLookup{Barcode: "036000291452", Name: "Should Not Be Used", Source: "web-fallback"}}
	service := &Service{provider: primary, webProvider: web, lookupCache: make(map[string]ProductLookup)}

	result, err := service.LookupExternalProduct(context.Background(), "036000291452")
	if err != nil || result == nil || result.Source != "upcitemdb" {
		t.Fatalf("expected primary result, got result=%+v err=%v", result, err)
	}
	if web.calls != 0 {
		t.Fatalf("web fallback was called after primary success: %d", web.calls)
	}
}

func TestServiceDoesNotCacheNoResultOrProviderError(t *testing.T) {
	for name, providerErr := range map[string]error{"no-result": errExternalNoResult, "provider-error": errors.New("provider unavailable")} {
		t.Run(name, func(t *testing.T) {
			primary := &stubBarcodeProvider{err: providerErr}
			web := &stubBarcodeProvider{err: errExternalNoResult}
			service := &Service{provider: primary, webProvider: web, lookupCache: make(map[string]ProductLookup)}
			_, err := service.LookupExternalProduct(context.Background(), "036000291452")
			if err == nil || len(service.lookupCache) != 0 {
				t.Fatalf("expected no cached result, err=%v cache=%v", err, service.lookupCache)
			}
		})
	}
}
