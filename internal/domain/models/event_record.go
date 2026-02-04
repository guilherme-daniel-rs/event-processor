package models

type EventRecord struct {
	ID            string `dynamodbav:"id"`
	EventID       string `dynamodbav:"event_id"`
	TenantID      string `dynamodbav:"tenant_id"`
	ClientID      string `dynamodbav:"client_id"`
	SchemaVersion string `dynamodbav:"schema_version"`
	OccurredAt    string `dynamodbav:"occurred_at"`
	Status        string `dynamodbav:"status"`
	Body          string `dynamodbav:"body"`
}

func (e EventRecord) TableName() *string {
	tableName := "events"
	return &tableName
}
