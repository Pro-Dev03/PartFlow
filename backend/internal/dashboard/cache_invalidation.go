package dashboard

// InvalidateDashboardCache invalidates the dashboard cache
// Call this function when any data that affects dashboard stats changes
func InvalidateDashboardCache() {
	if globalCacheService != nil {
		globalCacheService.InvalidateCache()
	}
}

// InvalidateDashboardCacheWithReason invalidates the dashboard cache with a reason for logging
// Call this function when any data that affects dashboard stats changes
func InvalidateDashboardCacheWithReason(reason string) {
	if globalCacheService != nil {
		globalCacheService.InvalidateCache()
		// Could add logging here if needed
	}
}