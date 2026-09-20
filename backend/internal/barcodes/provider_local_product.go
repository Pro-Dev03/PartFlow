package barcodes

import "context"

type localProductProvider struct {
	repo *Repository
}

func NewLocalProductProvider(repo *Repository) BarcodeProvider {
	return &localProductProvider{repo: repo}
}

func (p *localProductProvider) Lookup(ctx context.Context, barcode string) (*ProductLookup, error) {
	if p.repo == nil {
		return nil, errExternalNoResult
	}

	product, err := p.repo.GetProductByBarcode(ctx, barcode)
	if err != nil || product == nil || product.Name == "" {
		return nil, errExternalNoResult
	}

	return &ProductLookup{
		Barcode:    barcode,
		Name:       product.Name,
		Source:     "local-catalog",
		Confidence: 1,
	}, nil
}
