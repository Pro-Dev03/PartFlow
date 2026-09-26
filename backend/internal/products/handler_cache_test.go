package products

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestShouldCacheDefaultProductListOnlyWithoutFilters(t *testing.T) {
	base := httptest.NewRequest("GET", "/api/v1/products", nil)
	if !shouldCacheDefaultProductList(base, 1, 20) {
		t.Fatal("unfiltered first page should be cacheable")
	}

	for _, query := range []string{
		"search=brake", "category_id=category", "brand_id=brand", "track_serial=true",
		"track_individual=true", "low_stock_only=true", "in_stock_only=true",
		"manual_only=true", "sort_by=sku", "sort_order=DESC",
	} {
		t.Run(query, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/api/v1/products?"+query, nil)
			if shouldCacheDefaultProductList(request, 1, 20) {
				t.Fatalf("filtered product list must not use the default cache: %s", query)
			}
		})
	}
	if shouldCacheDefaultProductList(base, 2, 20) || shouldCacheDefaultProductList(base, 1, 100) {
		t.Fatal("non-default page or page size must not use the default cache")
	}
}

func TestProductCacheInvalidationPreventsStaleConcurrentWrite(t *testing.T) {
	cache := newProductsCache()
	revision := cache.currentRevision()
	cache.clear()
	cache.setAtRevision("stale", time.Minute, revision)
	if _, found := cache.get(); found {
		t.Fatal("an in-flight stale response must not repopulate the cache after invalidation")
	}
}
