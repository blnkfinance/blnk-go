package blnkgo

import "fmt"

// ListOptions pages a GET list endpoint through ListWithOptions (ledgers,
// balances, transactions, identities). A zero Limit or Offset is left out of
// the query, so Core's default page applies.
type ListOptions struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}

// listQuery validates the ListOptions passed to a ListWithOptions method and
// returns the value to hand NewRequest. It returns an untyped nil when there is
// nothing to send, so the request carries no query string.
func listQuery(o *ListOptions) (interface{}, error) {
	if o == nil {
		return nil, nil
	}
	if o.Limit < 0 {
		return nil, fmt.Errorf("invalid: limit must not be negative")
	}
	if o.Offset < 0 {
		return nil, fmt.Errorf("invalid: offset must not be negative")
	}
	if o.Limit == 0 && o.Offset == 0 {
		return nil, nil
	}
	return o, nil
}
