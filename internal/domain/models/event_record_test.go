package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventRecord_TableName(t *testing.T) {
	record := EventRecord{}
	tableName := record.TableName()
	assert.NotNil(t, tableName)
	assert.Equal(t, "events", *tableName)
}
