package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	blnkgo "github.com/blnkfinance/blnk-go"
)

func main() {
	baseURL, _ := url.Parse("http://localhost:5001/")
	client := blnkgo.NewClient(baseURL, nil, blnkgo.WithTimeout(
		5*time.Second,
	), blnkgo.WithRetry(2))

	file, _ := os.Open("file.csv")
	fmt.Println("file", file.Name())
	defer file.Close()
	reconUpload, resp, err := client.Reconciliation.Upload("stripe", file, file.Name())
	if err != nil {
		fmt.Print(err.Error())
		fmt.Println(resp)
		return
	}

	fmt.Println(reconUpload.UploadID)

	matchingRuleB := blnkgo.Matcher{
		Criteria: []blnkgo.Criteria{
			{
				Field:          "amount",
				Operator:       blnkgo.ReconciliationOperatorEquals,
				AllowableDrift: 0.1,
			},
		},
		Name: "Matching Rule",
	}

	matchingRule, resp, err := client.Reconciliation.CreateMatchingRule(matchingRuleB)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Print(resp.StatusCode)
	fmt.Printf("%+v\n", matchingRule)

	runReconBody := blnkgo.RunReconData{
		UploadID:        reconUpload.UploadID,
		MatchingRuleIDs: []string{matchingRule.RuleID},
		Strategy:        blnkgo.ReconciliationStrategyOneToMany,
		DryRun:          true,
	}

	runRecon, resp, err := client.Reconciliation.Run(runReconBody)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Print(resp.StatusCode)
	fmt.Println(runRecon.ReconciliationID)
}
