<p align="center">
  <img src="https://res.cloudinary.com/dmxizylxw/image/upload/v1724847576/blnk_github_logo_eyy2lf.png" alt="Blnk Logo" width="600"/>
</p>

# Blnk Go SDK Documentation

The official Go SDK for Blnk - A powerful ledger system for financial applications.

## Table of Contents

- [1. Installation](#1-installation)
  - [Prerequisites](#prerequisites)
  - [Step 1: Clone the Blnk Repository](#step-1-clone-the-blnk-repository)
  - [Step 2: Install Blnk Go SDK](#step-2-install-blnk-go-sdk)
  - [Step 3: Setting Up Configuration](#step-3-setting-up-configuration)
- [2. Launching Blnk](#2-launching-blnk)
- [3. Using the Blnk CLI](#3-using-the-blnk-cli)
- [4. Creating Your First Ledger](#4-creating-your-first-ledger)
- [5. Creating Balances](#5-creating-balances)
- [6. Recording Transactions](#6-recording-transactions)
- [7. Advanced Features](#7-advanced-features)
  - [Inflight Transactions](#inflight-transactions)
  - [Multi-Source/Destination Transactions](#multi-sourcedestination-transactions)
  - [Balance Monitors](#balance-monitors)
  - [Identity Management](#identity-management)
  - [Reconciliation](#reconciliation)
  - [Search](#search)
- [8. Examples](#8-examples)
- [Additional Resources](#additional-resources)

---

## 1. Installation

### Prerequisites

Ensure that you have the following installed on your machine:

- **Docker and Docker Compose** for running Blnk's server locally
- **Go 1.22 or later** for using the Blnk Go SDK

### Step 1: Clone the Blnk Repository

To start, clone the Blnk repository from GitHub:

```bash
git clone https://github.com/blnkfinance/blnk && cd blnk
```

### Step 2: Install Blnk Go SDK

Install the Blnk Go SDK in your project:

```bash
go get github.com/blnkfinance/blnk-go
```

### Step 3: Setting Up Configuration

In your cloned directory, create a configuration file named `blnk.json` with the following content:

```json
{
  "project_name": "Blnk",
  "data_source": {
    "dns": "postgres://postgres:password@postgres:5432/blnk?sslmode=disable"
  },
  "redis": {
    "dns": "redis:6379"
  },
  "server": {
    "domain": "blnk.io",
    "ssl": false,
    "ssl_email": "jerryenebeli@gmail.com",
    "port": "5001"
  },
  "notification": {
    "slack": {
      "webhook_url": "https://hooks.slack.com"
    }
  }
}
```

This configuration sets up connections to PostgreSQL and Redis, specifies your server details, and allows Slack notifications if needed.

---

## 2. Launching Blnk

With Docker Compose, launch the Blnk server:

```bash
docker compose up
```

Once running, your server will be accessible at [http://localhost:5001](http://localhost:5001/).

---

## 3. Using the Blnk CLI

The Blnk CLI offers quick access to manage ledgers, balances, and transactions. To verify the installation and view available commands, use:

```bash
blnk --help
```

---

## 4. Creating Your First Ledger

### What is a Ledger?

In Blnk, ledgers are used to categorize balances for organized tracking. When you first install Blnk, an internal ledger called the General Ledger is created by default.

### Step-by-Step: Creating a Ledger

Using the SDK, create a ledger for user accounts:

```go
package main

import (
    "fmt"
    "net/url"
    "time"
    
    blnkgo "github.com/blnkfinance/blnk-go"
)

func main() {
    // Initialize the Blnk client
    baseURL, _ := url.Parse("http://localhost:5001/")
    client := blnkgo.NewClient(
        baseURL, 
        nil, // API key (optional if not set on server)
        blnkgo.WithTimeout(5*time.Second),
        blnkgo.WithRetry(2),
    )
    
    // Create a new ledger
    ledgerBody := blnkgo.CreateLedgerRequest{
        Name: "Customer Savings Account",
        MetaData: map[string]interface{}{
            "project_owner": "YOUR_APP_NAME",
        },
    }
    
    newLedger, resp, err := client.Ledger.Create(ledgerBody)
    if err != nil {
        fmt.Printf("Error creating ledger: %v\n", err)
        return
    }
    
    fmt.Printf("Ledger Created: %+v\n", newLedger)
    fmt.Printf("Status Code: %d\n", resp.StatusCode)
}
```

This creates a new ledger for storing customer balances.

### Updating a Ledger Name

Rename an existing ledger without changing its ID or affecting balances and transactions:

```go
updateBody := blnkgo.UpdateLedgerRequest{
    Name: "Updated Customer Savings Account",
}

updatedLedger, resp, err := client.Ledger.Update(newLedger.LedgerID, updateBody)
if err != nil {
    fmt.Printf("Error updating ledger: %v\n", err)
    return
}

fmt.Printf("Ledger Updated: %+v\n", updatedLedger)
```

---

## 5. Creating Balances

Balances represent the store of value within a ledger, like a wallet or account. Each balance belongs to a ledger.

### Step-by-Step: Creating a Balance

To create a balance, specify the `ledger_id` and other details:

```go
balanceBody := blnkgo.CreateLedgerBalanceRequest{
    LedgerID: "ldg_073f7ffe-9dfd-42ce-aa50-d1dca1788adc",
    Currency: "USD",
    MetaData: map[string]interface{}{
        "first_name":     "Alice",
        "last_name":      "Hart",
        "account_number": "1234567890",
    },
}

newBalance, resp, err := client.LedgerBalance.Create(balanceBody)
if err != nil {
    fmt.Printf("Error creating balance: %v\n", err)
    return
}

fmt.Printf("Balance Created: %+v\n", newBalance)
```

### Viewing Balance Lineage

For balances with fund lineage tracking enabled, retrieve the provider breakdown (received, spent, available):

```go
lineage, resp, err := client.LedgerBalance.GetLineage(newBalance.BalanceID)
if err != nil {
    fmt.Printf("Error fetching balance lineage: %v\n", err)
    return
}

fmt.Printf("Balance Lineage: %+v\n", lineage)
fmt.Printf("Total with lineage: %s\n", lineage.TotalWithLineage.String())
```

### Linking an Identity to a Balance

Associate an existing identity with a balance:

```go
updateBody := blnkgo.UpdateBalanceIdentityRequest{
    IdentityID: "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6",
}

result, resp, err := client.LedgerBalance.UpdateIdentity(newBalance.BalanceID, updateBody)
if err != nil {
    fmt.Printf("Error updating balance identity: %v\n", err)
    return
}

fmt.Printf("Identity linked: %s\n", result.Message)
```

---

## 6. Recording Transactions

Transactions track financial activities within your application. Blnk ensures that each transaction is both immutable and idempotent.

### Step-by-Step: Recording a Transaction

To record a transaction, you'll need the `source` and `destination` balance IDs:

```go
transactionBody := blnkgo.CreateTransactionRequest{
    ParentTransaction: blnkgo.ParentTransaction{
        Amount:      750,
        Reference:   "ref_001adcfgf",
        Currency:    "USD",
        Precision:   100,
        Source:      "bln_28edb3e5-c168-4127-a1c4-16274e7a28d3",
        Destination: "bln_ebcd230f-6265-4d4a-a4ca-45974c47f746",
        Description: "Sent from app",
        MetaData: map[string]interface{}{
            "sender_name":    "John Doe",
            "sender_account": "00000000000",
        },
    },
}

newTransaction, resp, err := client.Transaction.Create(transactionBody)
if err != nil {
    fmt.Printf("Error recording transaction: %v\n", err)
    return
}

fmt.Printf("Transaction Recorded: %+v\n", newTransaction)
```

---

## 7. Advanced Features

### Inflight Transactions

Inflight transactions allow you to hold funds temporarily before committing or voiding them. This is useful for escrow scenarios, pending payments, or authorization holds.

#### Creating an Inflight Transaction

```go
inflightBody := blnkgo.CreateTransactionRequest{
    ParentTransaction: blnkgo.ParentTransaction{
        Amount:      1000,
        Reference:   "ref_inflight_001",
        Currency:    "USD",
        Precision:   100,
        Source:      "bln_source_id",
        Destination: "bln_destination_id",
        Description: "Escrow payment",
    },
    Inflight:           true,
    InflightExpiryDate: &expiryDate, // time.Time
}

transaction, resp, err := client.Transaction.Create(inflightBody)
```

#### Committing or Voiding an Inflight Transaction

```go
// Commit the transaction
updateBody := blnkgo.UpdateStatus{
    Status: blnkgo.InflightStatusCommit,
}

updatedTransaction, resp, err := client.Transaction.Update(
    "txn_id_here",
    updateBody,
)

// Or void the transaction
voidBody := blnkgo.UpdateStatus{
    Status: blnkgo.InflightStatusVoid,
}

voidedTransaction, resp, err := client.Transaction.Update(
    "txn_id_here",
    voidBody,
)
```

### Multi-Source/Destination Transactions

Split a transaction across multiple sources or destinations with custom distribution rules.

```go
multiSourceBody := blnkgo.CreateTransactionRequest{
    ParentTransaction: blnkgo.ParentTransaction{
        Amount:    1000,
        Reference: "ref_multi_001",
        Currency:  "USD",
        Precision: 100,
        Sources: []blnkgo.Source{
            {
                Identifier: "bln_source_1",
                Distribution: "50%",
                Narration:    "Split payment from source 1",
            },
            {
                Identifier: "bln_source_2",
                Distribution: "50%",
                Narration:    "Split payment from source 2",
            },
        },
        Destination: "bln_destination_id",
        Description: "Multi-source payment",
    },
}

transaction, resp, err := client.Transaction.Create(multiSourceBody)
```

### Balance Monitors

Set up monitors to track balance conditions and trigger webhooks when thresholds are met.

```go
monitorBody := blnkgo.MonitorData{
    BalanceID:   "bln_balance_id",
    Description: "Alert when balance falls below $100",
    Condition: blnkgo.MonitorCondition{
        Field:     "balance",
        Operator:  blnkgo.MonitorConditionOperatorLessThan,
        Value:     10000, // $100.00 with precision 100
        Precision: 100,
    },
    CallBackURL: "https://your-app.com/webhook/balance-alert",
}

monitor, resp, err := client.BalanceMonitor.Create(monitorBody)
if err != nil {
    fmt.Printf("Error creating monitor: %v\n", err)
    return
}

fmt.Printf("Monitor Created: %+v\n", monitor)
```

### Identity Management

Manage customer or organizational identities within your ledger system.

```go
identityBody := blnkgo.Identity{
    IdentityType: blnkgo.IdentityTypeIndividual,
    FirstName:    "John",
    LastName:     "Doe",
    EmailAddress: "john.doe@example.com",
    PhoneNumber:  "+1234567890",
    Nationality:  "US",
    Category:     "customer",
    Street:       "123 Main St",
    City:         "New York",
    State:        "NY",
    Country:      "USA",
    PostCode:     "10001",
    MetaData: map[string]interface{}{
        "customer_tier": "premium",
    },
}

identity, resp, err := client.Identity.Create(identityBody)
if err != nil {
    fmt.Printf("Error creating identity: %v\n", err)
    return
}

fmt.Printf("Identity Created: %+v\n", identity)
```

### Reconciliation

The reconciliation feature allows you to match and verify transactions against external data sources.

#### Upload Reconciliation Data

```go
// Upload CSV file for reconciliation
uploadResp, resp, err := client.Reconciliation.Upload(
    "path/to/file.csv",
    "external_source_name",
)
```

#### Create Matching Rules

```go
matcherBody := blnkgo.Matcher{
    Name:        "Amount and Reference Match",
    Description: "Match transactions by amount and reference",
    Criteria: []blnkgo.Criteria{
        {
            Field:    blnkgo.CriteriaFieldAmount,
            Operator: blnkgo.ReconciliationOperatorEquals,
        },
        {
            Field:    blnkgo.CriteriaFieldReference,
            Operator: blnkgo.ReconciliationOperatorEquals,
        },
    },
}

matchingRule, resp, err := client.Reconciliation.CreateMatchingRule(matcherBody)
```

#### Run Reconciliation

```go
reconBody := blnkgo.RunReconData{
    UploadID:         uploadResp.UploadID,
    Strategy:         blnkgo.ReconciliationStrategyOneToOne,
    DryRun:           false,
    GroupingCriteria: blnkgo.CriteriaFieldDate,
    MatchingRuleIDs:  []string{matchingRule.RuleID},
}

reconResp, resp, err := client.Reconciliation.RunReconciliation(reconBody)
```

### Search

Search across ledgers, balances, and transactions with flexible query parameters.

```go
searchParams := blnkgo.SearchParams{
    Q:        "john.doe@example.com",
    Page:     1,
    PerPage:  20,
}

// Search transactions
transactions, resp, err := client.Search.SearchTransactions(searchParams)

// Search balances
balances, resp, err := client.Search.SearchBalances(searchParams)

// Search ledgers
ledgers, resp, err := client.Search.SearchLedgers(searchParams)
```

---

## 8. Examples

The SDK includes several example applications demonstrating common use cases:

- **[Create Ledger](examples/create-ledger/)** - Basic ledger creation
- **[Escrow Application](examples/escrow-application/)** - Implementing an escrow system with inflight transactions
- **[Multi-Currency Wallet](examples/multi-currency-wallet/)** - Managing multiple currencies
- **[Savings Application](examples/savings-application/)** - Building a savings account system
- **[Virtual Card](examples/virtual-card/)** - Implementing virtual card functionality
- **[Balance Monitor](examples/balance-monitor/)** - Setting up balance monitoring and alerts
- **[Reconciliation](examples/reconciliation/)** - Transaction reconciliation workflows
- **[Multi-Sources](examples/multi-sources/)** - Multi-source/destination transactions

To run any example:

```bash
cd examples/<example-name>
go run main.go
```

---

## Additional Resources

For more examples and advanced use cases, please refer to the [Examples Code](https://github.com/blnkfinance/blnk-go/tree/main/examples).

### Issue Reporting

If you encounter any issues, please [report them on GitHub](https://github.com/blnkfinance/blnk/issues).
