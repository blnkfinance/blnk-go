package main

import (
	"fmt"
	"net/url"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
)

func main() {
	baseURL, _ := url.Parse("http://localhost:5001/")
	client := blnkgo.NewClient(baseURL, nil, blnkgo.WithTimeout(
		5*time.Second,
	), blnkgo.WithRetry(2))

	var ledgerBody blnkgo.CreateLedgerRequest = blnkgo.CreateLedgerRequest{
		Name: "First Ledger",
	}

	ledger, resp, err := client.Ledger.Create(ledgerBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	println(&resp.Body)
	println(ledger.LedgerID)
	println(ledger.Name)
	println(ledger.CreatedAt.GoString())

	ledgerList, resp, err := client.Ledger.List()
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	for _, el := range ledgerList {
		println(el.LedgerID)
		println(el.Name)
		println(el.CreatedAt.GoString())
	}
}
