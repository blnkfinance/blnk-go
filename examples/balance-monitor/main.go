package main

import (
	"fmt"
	"net/url"

	blnkgo "github.com/blnkfinance/blnk-go"
)

func main() {
	baseURL, _ := url.Parse("http://localhost:5001/")
	client := blnkgo.NewClient(baseURL, nil)

	ledgerBody := blnkgo.CreateLedgerRequest{
		Name: "Ledge",
		MetaData: blnkgo.MetaData{
			"project_name": "SendWorldApp",
		},
	}

	ledger, resp, err := client.Ledger.Create(ledgerBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(ledger.LedgerID)

	//create a balance
	ledgerBalanceBody := blnkgo.CreateLedgerBalanceRequest{
		LedgerID: ledger.LedgerID,
		Currency: "USD",
		MetaData: blnkgo.MetaData{
			"customer_name": "SendWorldApp",
		},
	}

	ledgerBalance, resp, err := client.LedgerBalance.Create(ledgerBalanceBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(ledgerBalance.BalanceID)

	//create a balance monitor
	ledgerBalanceMonitorBody := blnkgo.MonitorData{
		BalanceID: ledgerBalance.BalanceID,
		Condition: blnkgo.MonitorCondition{
			Field:     "credit_balance",
			Operator:  blnkgo.OperatorGreaterThan,
			Value:     1000,
			Precision: 100,
		},
	}

	bl, resp, err := client.BalanceMonitor.Create(ledgerBalanceMonitorBody)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(bl.BalanceID)
}
