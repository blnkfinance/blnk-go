package blnkgo_test

import (
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
	"github.com/stretchr/testify/require"
)

func TestValidateMonitorID(t *testing.T) {
	require.NoError(t, blnkgo.ValidateMonitorID("mon_test_123"))
	require.Error(t, blnkgo.ValidateMonitorID(""))
	require.Error(t, blnkgo.ValidateMonitorID("   "))
}
