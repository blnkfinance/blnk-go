package main

import (
	"fmt"
	"math/big"
	"net/url"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
)

func main() {
	baseURL, _ := url.Parse("http://localhost:5001/")
	client := blnkgo.NewClient(baseURL, nil, blnkgo.WithTimeout(
		5*time.Second,
	), blnkgo.WithRetry(2))

	_, _, err := client.Transaction.Create(blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			PreciseAmount: big.NewInt(10000),
			Reference:     "ref-21d",
			Precision:     100,
			Currency:      "USD",
			Description:   "Alice Funds",
			Destination:   "@alice",
			Sources: []blnkgo.Source{
				{
					Identifier:   "@test-1",
					Distribution: "5.500000%",
				},
				{
					Identifier:   "@test-2",
					Distribution: "left",
				},
			},
		},
	})
	fmt.Println(err)
}
