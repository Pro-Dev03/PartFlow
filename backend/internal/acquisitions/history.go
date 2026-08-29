package acquisitions

// ItemHistoryResponse contains the item snapshot and its immutable lifecycle
// events. Repair costs are included separately for the financials view.
type ItemHistoryResponse struct {
	Item             map[string]interface{}   `json:"item"`
	Events           []map[string]interface{} `json:"events"`
	RepairCosts      []map[string]interface{} `json:"repair_costs"`
	TotalRepairCosts float64                  `json:"total_repair_costs"`
}
