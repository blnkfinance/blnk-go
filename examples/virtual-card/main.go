package main

import (
	"log"
	"net/url"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
)

func main() {
	baseURL, _ := url.Parse("http://localhost:5001/")
	client := blnkgo.NewClient(baseURL, nil, blnkgo.WithTimeout(
		5*time.Second,
	), blnkgo.WithRetry(2))

	usdLedgerBody := blnkgo.CreateLedgerRequest{
		Name: "USD Ledger",
		MetaData: blnkgo.MetaData{
			"project_name": "USD virtual card",
		},
	}
	usdLedger, resp, err := client.Ledger.Create(usdLedgerBody)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("USD Ledger created: %+v\n", usdLedger)
	log.Printf("Response: %+v\n", resp)

	usdBalanceBody := blnkgo.CreateLedgerBalanceRequest{
		LedgerID: usdLedger.LedgerID,
		Currency: "USD",
		MetaData: blnkgo.MetaData{
			"customer_name":        "Jerry",
			"customer_internal_id": "1234",
			"card_state":           "ACTIVE",
			"card_number":          "411111XXXXXX1111", // Masked for security
			"card_expiry":          "12/26",
			"card_cvv":             "XXX", // Masked for security
		},
	}
	usdBalance, resp, err := client.LedgerBalance.Create(usdBalanceBody)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("USD Balance created: %+v\n", usdBalance)
	log.Printf("Response: %+v\n", resp)

	usdTransactionBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Currency:    "USD",
			Precision:   100,
			Reference:   "ref-05",
			Source:      "@World",
			Destination: "@Merchant",
			MetaData: blnkgo.MetaData{
				"merchant_name": "Store ABC",
				"customer_name": "Jerry",
			},
			Description: "Purchase at Store ABC",
		},
		AllowOverdraft: true,
	}
	usdTransaction, resp, err := client.Transaction.Create(usdTransactionBody)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("USD Transaction created: %+v\n", usdTransaction)
	log.Printf("Response: %+v\n", resp)

	inflightBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Currency:    "USD",
			Precision:   100,
			Reference:   "ref-06",
			Source:      "@Merchant",
			Destination: usdBalance.BalanceID,
			MetaData: blnkgo.MetaData{
				"merchant_name": "Store ABC",
				"customer_name": "Jerry",
			},
			Description: "Purchase at Store ABC",
		},
		Inflight: true,
	}
	inflightTransaction, resp, err := client.Transaction.Create(inflightBody)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Inflight Transaction created: %+v\n", inflightTransaction)
	log.Printf("Response: %+v\n", resp)
	///sleep for 4 seconds to simulate waiting for a webhook or action to commit the transaction, also allows for it to be processed by the background job
	time.Sleep(4 * time.Second)

	_, _, err = client.Transaction.Update(inflightTransaction.TransactionID, blnkgo.UpdateStatus{
		Status: blnkgo.InflightStatusCommit,
	})

	if err != nil {
		log.Fatal(err)
	}
}
