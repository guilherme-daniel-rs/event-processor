package models

type EventRecord struct {
	ID            string
	EventID       string
	TenantID      string
	ClientID      string
	SchemaVersion string
	OccurredAt    string
	Status        string
	Body          string
}

func (e EventRecord) TableName() *string {
	tableName := "events"
	return &tableName
}
