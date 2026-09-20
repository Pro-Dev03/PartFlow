package barcodes

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type UPCItemDBProvider struct{}

func NewUPCItemDBProvider() BarcodeProvider {
	return &UPCItemDBProvider{}
}

func (p *UPCItemDBProvider) Lookup(ctx context.Context, barcode string) (*ProductLookup, error) {
	var payload upcItemDBResponse
	endpoint := "https://api.upcitemdb.com/prod/trial/lookup?upc=" + url.QueryEscape(barcode)
	if err := externalJSONRequest(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	if len(payload.Items) == 0 {
		return nil, errExternalNoResult
	}

	item := payload.Items[0]
	image := item.Image
	if image == "" && len(item.Images) > 0 {
		image = item.Images[0]
	}
	name := strings.TrimSpace(item.Title)
	if name == "" {
		name = strings.TrimSpace(item.Model)
	}
	if name == "" {
		return nil, fmt.Errorf("UPCitemDB result has no product name")
	}

	return &ProductLookup{
		Barcode:     barcode,
		Name:        strings.TrimSpace(name),
		Brand:       strings.TrimSpace(item.Brand),
		Category:    strings.TrimSpace(item.Category),
		Description: strings.TrimSpace(item.Description),
		Size:        strings.TrimSpace(item.Size),
		Image:       strings.TrimSpace(image),
		Source:      "upcitemdb",
		Confidence:  0.85,
	}, nil
}
