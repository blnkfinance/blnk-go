package blnkgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStartReindexRequest(t *testing.T) {
	assert.NoError(t, ValidateStartReindexRequest(StartReindexRequest{}))
	batchSize := 1000
	assert.NoError(t, ValidateStartReindexRequest(StartReindexRequest{BatchSize: &batchSize}))

	zero := 0
	err := ValidateStartReindexRequest(StartReindexRequest{BatchSize: &zero})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch_size must be a positive integer")

	negative := -1
	err = ValidateStartReindexRequest(StartReindexRequest{BatchSize: &negative})
	assert.Error(t, err)
}
