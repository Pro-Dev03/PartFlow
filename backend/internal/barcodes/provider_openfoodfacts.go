package barcodes

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type openFoodFactsProvider struct{}

type openFoodFactsResponse struct {
	Status  int `json:"status"`
	Product struct {
		Code          string `json:"code"`
		ProductName   string `json:"product_name"`
		GenericName   string `json:"generic_name"`
		Brands        string `json:"brands"`
		Categories    string `json:"categories"`
		Ingredients   string `json:"ingredients_text"`
		ImageFrontURL string `json:"image_front_url"`
	} `json:"product"`
}

func NewOpenFoodFactsProvider() BarcodeProvider {
	return &openFoodFactsProvider{}
}

func (p *openFoodFactsProvider) Lookup(ctx context.Context, barcode string) (*ProductLookup, error) {
	endpoint := "https://world.openfoodfacts.org/api/v2/product/" + url.PathEscape(barcode) + ".json"
	var payload openFoodFactsResponse
	if err := externalJSONRequest(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	if payload.Status != 1 || !barcodeMatches(payload.Product.Code, barcode) {
		return nil, errExternalNoResult
	}

	name := strings.TrimSpace(payload.Product.ProductName)
	if name == "" {
		name = strings.TrimSpace(payload.Product.GenericName)
	}
	if name == "" {
		return nil, fmt.Errorf("Open Food Facts result has no product name")
	}

	return &ProductLookup{
		Barcode:     barcode,
		Name:        name,
		Brand:       strings.TrimSpace(payload.Product.Brands),
		Category:    strings.TrimSpace(payload.Product.Categories),
		Description: strings.TrimSpace(payload.Product.Ingredients),
		Image:       strings.TrimSpace(payload.Product.ImageFrontURL),
		Source:      "openfoodfacts",
		Confidence:  0.8,
	}, nil
}

func barcodeMatches(found, requested string) bool {
	found = strings.TrimSpace(found)
	return found == requested || (len(requested) == 12 && found == "0"+requested)
}
