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

	var LedgerBody blnkgo.CreateLedgerRequest = blnkgo.CreateLedgerRequest{
		Name: "USD Ledger",
	}

	esrowLedger, resp, err := client.Ledger.Create(LedgerBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println(resp.StatusCode)
	fmt.Printf("%+v\n", esrowLedger)

	escrowBalanceBody := blnkgo.CreateLedgerBalanceRequest{
		LedgerID: esrowLedger.LedgerID,
		Currency: "USD",
		MetaData: blnkgo.MetaData{
			"account_type":        "Escrow",
			"customer_name":       "Alice Johnson",
			"customer_id":         "CUST001",
			"account_opened_date": "2024-01-01",
			"account_status":      "active",
		},
	}

	escrowBalance, resp, err := client.LedgerBalance.Create(escrowBalanceBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Println(resp.StatusCode)

	escrowBalanceBody2 := blnkgo.CreateLedgerBalanceRequest{
		LedgerID: esrowLedger.LedgerID,
		Currency: "USD",
		MetaData: blnkgo.MetaData{
			"account_type":        "Escrow",
			"customer_name":       "Bob Smith",
			"customer_id":         "CUST002",
			"account_opened_date": "2024-01-01",
			"account_status":      "active",
		},
	}

	escrowBalance2, resp, err := client.LedgerBalance.Create(escrowBalanceBody2)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Println(resp.StatusCode)
	fundAliceBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-21",
			Precision:   100,
			Currency:    "USD",
			Source:      "@bank-account",
			Destination: escrowBalance.BalanceID,
			MetaData: blnkgo.MetaData{
				"transaction_type": "deposit",
				"customer_name":    "Alice Johnson",
				"customer_id":      "alice-5786",
			},
			Description: "Alice Funds",
		},
		Inflight: true,
	}

	fundAlice, resp, err := client.Transaction.Create(fundAliceBody)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Printf("%+v\n", fundAlice)
	fmt.Println(resp.StatusCode)

	fundBobBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-22",
			Precision:   100,
			Currency:    "USD",
			Source:      escrowBalance.BalanceID,
			Destination: escrowBalance2.BalanceID,
			MetaData: blnkgo.MetaData{
				"transaction_type": "release",
				"customer_name":    "Bob Smith",
				"customer_id":      "bob-5786",
			},
			Description: "Fund Bob",
		},
	}

	fundBob, _, err := client.Transaction.Create(fundBobBody)

	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("%+v\n", fundBob.TransactionID)

	//refunding Alice
	refundAliceBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-23",
			Precision:   100,
			Currency:    "USD",
			Source:      escrowBalance2.BalanceID,
			Destination: escrowBalance.BalanceID,
			MetaData: blnkgo.MetaData{
				"transaction_type": "refund",
				"customer_name":    "Alice Johnson",
				"customer_id":      "alice-5786",
			},
			Description: "Alice refund",
		},
	}

	refund, resp, err := client.Transaction.Create(refundAliceBody)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("%+v\n", refund)
	fmt.Println(resp.StatusCode)
}
