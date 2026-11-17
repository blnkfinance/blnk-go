package blnkgo_test

import (
	"net/http"
	"net/url"
	"testing"

	blnkgo "github.com/blnkfinance/blnk-go"
)

// These interfaces copy the v1.6.0 signatures that callers may have declared
// in their own code (for mocks or adapters). If a method's signature changes,
// this file stops compiling, so a source break is caught before release.
type (
	v160LedgerLister interface {
		List() ([]blnkgo.Ledger, *http.Response, error)
	}
	v160BalanceLister interface {
		List() ([]blnkgo.LedgerBalance, *http.Response, error)
	}
	v160TransactionLister interface {
		List() ([]blnkgo.Transaction, *http.Response, error)
	}
	v160IdentityLister interface {
		List() ([]*blnkgo.IdentityResponse, *http.Response, error)
	}
	v160MonitorUpdater interface {
		Update(monitorID string, data blnkgo.MonitorData) (*blnkgo.MonitorDataResp, *http.Response, error)
	}
	v160MetadataUpdater interface {
		UpdateMetadata(entityID string, body blnkgo.UpdateMetaDataRequest) (*blnkgo.Metadata, *http.Response, error)
	}
	v160InflightUpdater interface {
		Update(transactionID string, body blnkgo.UpdateStatus) (*blnkgo.Transaction, *http.Response, error)
	}
	v160BalanceGetter interface {
		Get(balanceID string, opts ...*blnkgo.GetBalanceRequest) (*blnkgo.LedgerBalance, *http.Response, error)
	}
	v160Filterer interface {
		Filter(params blnkgo.FilterParams) (*blnkgo.FilterResponse, *http.Response, error)
	}
)

var (
	_ v160LedgerLister      = (*blnkgo.LedgerService)(nil)
	_ v160BalanceLister     = (*blnkgo.LedgerBalanceService)(nil)
	_ v160TransactionLister = (*blnkgo.TransactionService)(nil)
	_ v160IdentityLister    = (*blnkgo.IdentityService)(nil)
	_ v160MonitorUpdater    = (*blnkgo.BalanceMonitorService)(nil)
	_ v160MetadataUpdater   = (*blnkgo.MetadataService)(nil)
	_ v160InflightUpdater   = (*blnkgo.TransactionService)(nil)
	_ v160BalanceGetter     = (*blnkgo.LedgerBalanceService)(nil)
	_ v160Filterer          = (*blnkgo.TransactionService)(nil)
	_ v160Filterer          = (*blnkgo.LedgerBalanceService)(nil)
	_ v160Filterer          = (*blnkgo.LedgerService)(nil)
	_ v160Filterer          = (*blnkgo.IdentityService)(nil)
)

// Method values keep their v1.6.0 function types.
func TestV160MethodValueTypes(t *testing.T) {
	base, _ := url.Parse("http://localhost:5001")
	c := blnkgo.NewClient(base, nil)
	var (
		_ func() ([]blnkgo.Ledger, *http.Response, error)            = c.Ledger.List
		_ func() ([]blnkgo.LedgerBalance, *http.Response, error)     = c.LedgerBalance.List
		_ func() ([]blnkgo.Transaction, *http.Response, error)       = c.Transaction.List
		_ func() ([]*blnkgo.IdentityResponse, *http.Response, error) = c.Identity.List
	)
}
