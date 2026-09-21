package archive

type Event struct {
	ID            string   `db:"id" json:"id"`
	EventType     string   `db:"event_type" json:"event_type"`
	Section       string   `db:"section" json:"section"`
	EntityType    string   `db:"entity_type" json:"entity_type"`
	EntityID      string   `db:"entity_id" json:"entity_id"`
	ReferenceType string   `db:"reference_type" json:"reference_type,omitempty"`
	ReferenceID   string   `db:"reference_id" json:"reference_id,omitempty"`
	UserID        string   `db:"user_id" json:"user_id,omitempty"`
	UserName      string   `db:"user_name" json:"user_name,omitempty"`
	Status        string   `db:"status" json:"status"`
	Description   string   `db:"description" json:"description,omitempty"`
	Details       string   `db:"details" json:"details,omitempty"`
	Quantity      *float64 `db:"quantity" json:"quantity,omitempty"`
	BeforeValue   *float64 `db:"before_value" json:"before_value,omitempty"`
	AfterValue    *float64 `db:"after_value" json:"after_value,omitempty"`
	CreatedAt     string   `db:"created_at" json:"created_at"`
}

type ListRequest struct {
	Page       int
	PerPage    int
	Search     string
	Section    string
	EventType  string
	UserID     string
	Status     string
	EntityType string
	EntityID   string
	Reference  string
	StartDate  string
	EndDate    string
	SortBy     string
	SortOrder  string
}
