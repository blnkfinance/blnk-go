package blnkgo

import "fmt"

// ListOptions pages a GET list endpoint (ledgers, balances, transactions,
// identities). A zero Limit or Offset is left out of the query, so Core's
// default page applies.
type ListOptions struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`
}

// listQuery validates the optional ListOptions passed to a List method and
// returns the value to hand NewRequest. It returns an untyped nil when no
// options were given so the request carries no query string.
func listQuery(opts []*ListOptions) (interface{}, error) {
	if len(opts) > 1 {
		return nil, fmt.Errorf("List accepts at most one ListOptions")
	}
	if len(opts) == 0 || opts[0] == nil {
		return nil, nil
	}
	o := opts[0]
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
