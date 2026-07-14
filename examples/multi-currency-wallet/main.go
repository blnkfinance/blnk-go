package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
)

type LedgerBalanceResponseCh struct {
	Resp *http.Response
	Data *blnkgo.LedgerBalance
	Err  error
}

func main() {
	baseURL, _ := url.Parse("http://localhost:5001/")
	client := blnkgo.NewClient(baseURL, nil, blnkgo.WithTimeout(
		5*time.Second,
	), blnkgo.WithRetry(2))

	var usdLedgerBody blnkgo.CreateLedgerRequest = blnkgo.CreateLedgerRequest{
		Name: "USD Ledger",
	}

	usdLedger, resp, err := client.Ledger.Create(usdLedgerBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println(resp.StatusCode)
	fmt.Printf("%+v\n", usdLedger)

	var eurLedgerBody blnkgo.CreateLedgerRequest = blnkgo.CreateLedgerRequest{
		Name: "EUR Ledger",
	}

	eurLedger, resp, err := client.Ledger.Create(eurLedgerBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println(resp.StatusCode)
	fmt.Printf("%+v\n", eurLedger)

	usdBalanceBody := blnkgo.CreateLedgerBalanceRequest{
		LedgerID: usdLedger.LedgerID,
		Currency: "USD",
	}

	eurBalanceBody := blnkgo.CreateLedgerBalanceRequest{
		LedgerID: eurLedger.LedgerID,
		Currency: "EUR",
	}

	//use concurrency to create both ledger balances
	usdBalanceChan := make(chan LedgerBalanceResponseCh)
	eurBalanceChan := make(chan LedgerBalanceResponseCh)

	go func() {
		usdBalance, resp, err := client.LedgerBalance.Create(usdBalanceBody)
		if err != nil {
			fmt.Print(err.Error())
			return
		}

		fmt.Println(resp.StatusCode)
		usdBalanceChan <- LedgerBalanceResponseCh{
			Resp: resp,
			Data: usdBalance,
			Err:  err,
		}
	}()

	go func() {
		eurBalance, resp, err := client.LedgerBalance.Create(eurBalanceBody)
		if err != nil {
			fmt.Print(err.Error())
			return
		}

		fmt.Println(resp.StatusCode)
		eurBalanceChan <- LedgerBalanceResponseCh{
			Resp: resp,
			Data: eurBalance,
			Err:  err,
		}
	}()
	usdBalanceResp := <-usdBalanceChan
	eurBalanceResp := <-eurBalanceChan

	if usdBalanceResp.Err != nil || eurBalanceResp.Err != nil {
		fmt.Print("Error creating ledger balances")
		fmt.Print(usdBalanceResp.Err.Error())
		fmt.Print(eurBalanceResp.Err.Error())
		return
	}
	usdBalance := usdBalanceResp.Data
	eurBalance := eurBalanceResp.Data

	fmt.Printf("%+v\n", *usdBalance)
	fmt.Printf("%+v\n", *eurBalance)
	//create transaction
	transactionBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-01",
			Precision:   100,
			Currency:    "USD",
			Source:      "@World",
			Destination: usdBalance.BalanceID,
			Description: "Usd Exchange",
		},
		AllowOverdraft: true,
	}

	transaction, resp, err := client.Transaction.Create(transactionBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println(resp.StatusCode)
	fmt.Printf("%+v\n", transaction)

	eurTransactionBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      1000,
			Reference:   "ref-02",
			Precision:   100,
			Currency:    "EUR",
			Source:      "@World",
			Destination: eurBalance.BalanceID,
			Description: "Eur Exchange",
		},
		AllowOverdraft: true,
	}

	fmt.Printf("%+v\n", eurTransactionBody)
	_, resp, err = client.Transaction.Create(eurTransactionBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Println(resp.StatusCode)
	//simulate waiting for the transactions to complete on the service by sleeping
	time.Sleep(5 * time.Second)

	//create a debit on usd balance by making it the source and destination the world
	debitBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      100,
			Reference:   "ref-03",
			Precision:   100,
			Currency:    "USD",
			Source:      usdBalance.BalanceID,
			Destination: "@World",
			Description: "Debit",
		},
	}

	debit, resp, err := client.Transaction.Create(debitBody)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Printf("Debit: %+v\n", debit)
	fmt.Println(resp.StatusCode)

	//move money from the eur balance to the usd balance and set a rate
	exchangeBody := blnkgo.CreateTransactionRequest{
		ParentTransaction: blnkgo.ParentTransaction{
			Amount:      100,
			Reference:   "ref-04",
			Precision:   100,
			Currency:    "EUR",
			Source:      eurBalance.BalanceID,
			Destination: usdBalance.BalanceID,
			Rate:        1.1,
			Description: "Exchange",
		},
	}
	fmt.Printf("%+v\n", exchangeBody)
	exchange, resp, err := client.Transaction.Create(exchangeBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("Exchange: %+v\n", exchange)
	fmt.Printf("Exchange: %+v\n", resp)
	//simulate waiting for the transactions to complete on the service by sleeping
	time.Sleep(5 * time.Second)

	//get the balance of the usd balance
	usdBalance, resp, err = client.LedgerBalance.Get(usdBalance.BalanceID)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("USD Balance: %+v\n", usdBalance)
	fmt.Println(resp.StatusCode)

	//get the balance of the eur balance
	eurBalance, resp, err = client.LedgerBalance.Get(eurBalance.BalanceID)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("EUR Balance: %+v\n", eurBalance)
	fmt.Println(resp.StatusCode)
}
