package barcodes

import "context"

type ProductLookup = ExternalProductLookup

type BarcodeProvider interface {
	Lookup(ctx context.Context, barcode string) (*ProductLookup, error)
}
