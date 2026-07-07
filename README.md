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
  - [Error Handling](#error-handling)
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
        MetaData: blnkgo.MetaData{
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

### Retries

`WithRetry` sets the **total number of attempts** (first try included). Default is `1` (no retries).

```go
client := blnkgo.NewClient(
    baseURL,
    &apiKey,
    blnkgo.WithRetry(3),                    // up to 3 attempts
    blnkgo.WithRetryDelay(2 * time.Second), // linear backoff base delay
)
```

Retry behavior (aligned with the TypeScript SDK):

- **GET** requests retry on **5xx** responses and retryable network errors
- **POST**, **PUT**, and **DELETE** are **not** retried (avoids duplicate money movement)
- Request timeouts are not retried
- Backoff delay is `RetryDelay × attempt` between retries (2s, 4s, … with default delay)

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
    MetaData: blnkgo.MetaData{
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

To enable fund lineage tracking on create, set `track_fund_lineage` (requires `identity_id`) and optionally `allocation_strategy`:

```go
balanceBody := blnkgo.CreateLedgerBalanceRequest{
    LedgerID:           "ldg_073f7ffe-9dfd-42ce-aa50-d1dca1788adc",
    IdentityID:         "idt_3b63c8da-af29-4cc3-ad38-df17d87456e6",
    Currency:           "USD",
    TrackFundLineage:   true,
    AllocationStrategy: blnkgo.AllocationStrategyFIFO, // FIFO | LIFO | PROPORTIONAL
}

lineageBalance, resp, err := client.LedgerBalance.Create(balanceBody)
```

### Retrieving a Balance (from source)

By default, `Get` returns the stored balance snapshot. Pass `from_source: true` to reconstruct the balance from all transactions instead:

```go
balance, resp, err := client.LedgerBalance.Get(newBalance.BalanceID, &blnkgo.GetBalanceRequest{
    FromSource: true,
})
if err != nil {
    fmt.Printf("Error fetching balance from source: %v\n", err)
    return
}

fmt.Printf("Balance from source: %+v\n", balance)
```

Existing callers can keep using `client.LedgerBalance.Get(balanceID)` with no second argument.

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

### Creating Balance Snapshots

Capture daily balance snapshots in batches. Optionally set `BatchSize` to control how many balances are processed per batch (server default is 1000):

```go
snapshotBody := blnkgo.CreateBalanceSnapshotRequest{
    BatchSize: 500,
}

result, resp, err := client.LedgerBalance.CreateSnapshot(snapshotBody)
if err != nil {
    fmt.Printf("Error creating balance snapshot: %v\n", err)
    return
}

fmt.Printf("Snapshot started: %s\n", result.Message)
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
        MetaData: blnkgo.MetaData{
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

### Recording Bulk Transactions

Submit multiple transactions in a single request (up to `MaxBulkCreateItems`, 10,000 per request). Set `Atomic` to ensure all transactions succeed or fail together:

```go
bulkBody := blnkgo.CreateBulkTransactionRequest{
    Atomic: true,
    Transactions: []blnkgo.CreateTransactionRequest{
        {
            ParentTransaction: blnkgo.ParentTransaction{
                Amount:      500,
                Reference:   "bulk_ref_001",
                Currency:    "USD",
                Precision:   100,
                Source:      "bln_source_id",
                Destination: "bln_destination_id",
                Description: "Bulk payment 1",
            },
        },
        {
            ParentTransaction: blnkgo.ParentTransaction{
                Amount:      750,
                Reference:   "bulk_ref_002",
                Currency:    "USD",
                Precision:   100,
                Source:      "bln_source_id",
                Destination: "bln_destination_id",
                Description: "Bulk payment 2",
            },
        },
    },
}

bulkResult, resp, err := client.Transaction.CreateBulk(bulkBody)
if err != nil {
    fmt.Printf("Error creating bulk transactions: %v\n", err)
    return
}

fmt.Printf("Bulk batch %s: %d transactions (%s)\n",
    bulkResult.BatchID, bulkResult.TransactionCount, bulkResult.Status)
```

### Recovering Stuck Queued Transactions

Manually trigger recovery of transactions stuck in the queue (`POST /transactions/recover`). Optionally pass a `threshold` duration (e.g. `5m`, `1h`):

```go
result, resp, err := client.Transaction.RecoverQueue(blnkgo.RecoverQueueRequest{
    Threshold: "5m",
})
if err != nil {
    fmt.Printf("Error recovering queue: %v\n", err)
    return
}

fmt.Printf("Recovered %d transactions (threshold %s)\n", result.Recovered, result.Threshold)
```

### Refunding a Transaction

Refund by transaction ID. Omit the body to queue the refund (Core default), or pass `skip_queue: true` for synchronous processing:

```go
// Queued refund (default) — existing callers keep working without a second argument
refund, resp, err := client.Transaction.Refund(originalTxnID)

// Synchronous refund
refund, resp, err := client.Transaction.Refund(originalTxnID, &blnkgo.RefundTransactionRequest{
    SkipQueue: true,
})
```

### Getting a Transaction by Reference

Look up a transaction using its unique reference string:

```go
transaction, resp, err := client.Transaction.GetByReference("ref_8d2ce2f0-0d75-4a91-9d43-2ad2c2e6b9ad")
if err != nil {
    fmt.Printf("Error fetching transaction by reference: %v\n", err)
    return
}

fmt.Printf("Transaction %s: %+v\n", transaction.TransactionID, transaction)
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
// Commit the transaction (queued by default on Core 0.15.0+)
updateBody := blnkgo.UpdateStatus{
    Status: blnkgo.InflightStatusCommit,
}

updatedTransaction, resp, err := client.Transaction.Update(
    "txn_id_here",
    updateBody,
)

// Synchronous commit (skip_queue: true) — returns APPLIED immediately
syncCommitBody := blnkgo.UpdateStatus{
    Status:    blnkgo.InflightStatusCommit,
    SkipQueue: true,
}

syncCommitted, resp, err := client.Transaction.Update(
    "txn_id_here",
    syncCommitBody,
)

// Or void the transaction
voidBody := blnkgo.UpdateStatus{
    Status: blnkgo.InflightStatusVoid,
}

voidedTransaction, resp, err := client.Transaction.Update(
    "txn_id_here",
    voidBody,
)

// Synchronous void
syncVoidBody := blnkgo.UpdateStatus{
    Status:    blnkgo.InflightStatusVoid,
    SkipQueue: true,
}
```

#### Bulk Commit Inflight Transactions

Commit multiple independently-created inflight transactions in a single call:

```go
bulkCommitBody := blnkgo.BulkCommitInflightRequest{
    SkipQueue: true, // optional: process synchronously without queuing
    Transactions: []blnkgo.BulkCommitInflightItem{
        {TransactionID: "txn_id_1"},
        {TransactionID: "txn_id_2", Amount: 40},
        {TransactionID: "txn_id_3", PreciseAmount: big.NewInt(125034)},
    },
}

bulkResult, resp, err := client.Transaction.BulkCommitInflight(bulkCommitBody)
if err != nil {
    fmt.Printf("Error bulk committing inflight transactions: %v\n", err)
    return
}

fmt.Printf("Bulk commit: %d succeeded, %d failed\n", bulkResult.Succeeded, bulkResult.Failed)
for _, r := range bulkResult.Results {
    fmt.Printf("  %s: %s\n", r.TransactionID, r.Status)
}
```

#### Bulk Void Inflight Transactions

Void multiple independently-created inflight transactions in a single call:

```go
bulkVoidBody := blnkgo.BulkVoidInflightRequest{
    SkipQueue: true, // optional: process synchronously without queuing
    TransactionIDs: []string{
        "txn_id_1",
        "txn_id_2",
        "txn_id_3",
    },
}

bulkResult, resp, err := client.Transaction.BulkVoidInflight(bulkVoidBody)
if err != nil {
    fmt.Printf("Error bulk voiding inflight transactions: %v\n", err)
    return
}

fmt.Printf("Bulk void: %d succeeded, %d failed\n", bulkResult.Succeeded, bulkResult.Failed)
for _, r := range bulkResult.Results {
    fmt.Printf("  %s: %s\n", r.TransactionID, r.Status)
}
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

Delete a balance monitor (Core 0.15.0+):

```go
deleted, resp, err := client.BalanceMonitor.Delete(monitor.MonitorID)
if err != nil {
    log.Fatal(err)
}
fmt.Println(deleted.Message) // BalanceMonitor deleted successfully
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
    MetaData: blnkgo.MetaData{
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

To supply a deterministic `identity_id` (must be `idt_` + UUID), set it on the request body:

```go
identityBody := blnkgo.Identity{
    IdentityID:   "idt_8c5a8e2f-3f1d-5a9b-9c3e-4d8f1e5a7b2c",
    IdentityType: blnkgo.Individual,
    FirstName:    "John",
    Category:     "customer",
}

identity, resp, err := client.Identity.Create(identityBody)
```

Only `identity_type` (when set) and `identity_id` format are validated client-side; other fields are optional per the API reference.

Filter identities with structured query filters:

```go
result, resp, err := client.Identity.Filter(blnkgo.FilterParams{
    Filters: []blnkgo.Filter{
        {Field: "email_address", Operator: blnkgo.OpEqual, Value: "john.doe@example.com"},
        {Field: "category", Operator: blnkgo.OpEqual, Value: "customer"},
    },
    Limit:        20,
    Offset:       0,
    IncludeCount: true,
})
if err != nil {
    fmt.Printf("Error filtering identities: %v\n", err)
    return
}

fmt.Printf("Filter status: %d\n", resp.StatusCode)
if result.TotalCount != nil {
    fmt.Printf("Total matches: %d\n", *result.TotalCount)
}
fmt.Printf("Data: %+v\n", result.Data)
```

Tokenize a single PII field (use PascalCase struct field names such as `FirstName`, not `first_name`):

```go
tokenized, resp, err := client.Identity.TokenizeField(
    identity.IdentityId,
    string(blnkgo.TokenizableFieldFirstName),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(tokenized.Message) // Field tokenized successfully
```

Tokenize multiple PII fields in one request:

```go
tokenized, resp, err := client.Identity.Tokenize(identity.IdentityId, blnkgo.TokenizeRequest{
    Fields: []blnkgo.TokenizableIdentityField{
        blnkgo.TokenizableFieldFirstName,
        blnkgo.TokenizableFieldLastName,
        blnkgo.TokenizableFieldEmailAddress,
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(tokenized.Message) // Fields tokenized successfully
```

List fields currently tokenized on an identity:

```go
fields, resp, err := client.Identity.GetTokenizedFields(identity.IdentityId)
if err != nil {
    log.Fatal(err)
}
for _, field := range fields.TokenizedFields {
    fmt.Println(field) // e.g. FirstName, EmailAddress
}
```

Detokenize a single field and read the original value:

```go
detokenized, resp, err := client.Identity.DetokenizeField(
    identity.IdentityId,
    string(blnkgo.TokenizableFieldEmailAddress),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(detokenized.Field, detokenized.Value)
```

Detokenize multiple fields (pass an empty `fields` slice to detokenize all tokenized fields):

```go
detokenized, resp, err := client.Identity.Detokenize(identity.IdentityId, blnkgo.DetokenizeRequest{
    Fields: []blnkgo.TokenizableIdentityField{
        blnkgo.TokenizableFieldFirstName,
        blnkgo.TokenizableFieldEmailAddress,
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(detokenized.Fields["FirstName"], detokenized.Fields["EmailAddress"])
```

Delete an identity (Core 0.15.0+):

```go
deleted, resp, err := client.Identity.Delete(identity.IdentityId)
if err != nil {
    log.Fatal(err)
}
fmt.Println(deleted.Message) // Identity deleted successfully
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

#### Update Matching Rule

```go
updatedRule, resp, err := client.Reconciliation.UpdateMatchingRule(matchingRule.RuleID, blnkgo.Matcher{
    Name:        "Updated Amount Match",
    Description: "Updated match by amount only",
    Criteria: []blnkgo.Criteria{
        {
            Field:    blnkgo.CriteriaFieldAmount,
            Operator: blnkgo.ReconciliationOperatorEquals,
        },
    },
})
```

#### Delete Matching Rule

```go
deleted, resp, err := client.Reconciliation.DeleteMatchingRule(matchingRule.RuleID)
if err != nil {
    panic(err)
}
fmt.Println(deleted.Message)
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

reconResp, resp, err := client.Reconciliation.Run(reconBody)
if err != nil {
    log.Fatal(err)
}
// reconResp.ReconciliationID — use webhooks or GET /reconciliation/{id} for results (Core 0.15.0+)
fmt.Println(reconResp.ReconciliationID)
```

#### Run Instant Reconciliation

Reconcile inline external transactions without uploading a file first.

```go
instantResp, resp, err := client.Reconciliation.RunInstant(blnkgo.RunInstantReconData{
    ExternalTransactions: []blnkgo.ExternalTransaction{
        {
            ID:          "txn1a2b3c4d5e6f7g8h9i0",
            Amount:      5.49,
            Reference:   "INV-2023-002",
            Currency:    "GBP",
            Description: "Card payment",
            Date:        func() *time.Time { t := time.Date(2024, 11, 15, 14, 25, 30, 0, time.UTC); return &t }(),
            Source:      "bank-api",
        },
    },
    Strategy:        blnkgo.ReconciliationStrategyOneToOne,
    DryRun:          true,
    MatchingRuleIDs: []string{matchingRule.RuleID},
})
// instantResp.ReconciliationID — poll GET /reconciliation/{id} for status
```

#### Get Reconciliation

Poll reconciliation status and match counts after starting a run.

```go
recon, resp, err := client.Reconciliation.Get(instantResp.ReconciliationID)
if err != nil {
    panic(err)
}
fmt.Println("Status:", recon.Status)
fmt.Println("Matched:", recon.MatchedTransactions)
fmt.Println("Unmatched:", recon.UnmatchedTransactions)
```

### Health

Check whether Blnk Core is running and reachable:

```go
health, resp, err := client.Health.Check()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Status:", health.Status) // UP when Core is healthy
```

### API keys

Create a scoped API key (requires master key or `api-keys:write` scope):

```go
apiKey, resp, err := client.ApiKeys.Create(blnkgo.CreateApiKeyRequest{
    Name:      "read-only-ledger",
    Owner:     "team_acme",
    Scopes:    []string{"ledgers:read"},
    ExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Key ID:", apiKey.ApiKeyID)
fmt.Println("Secret:", apiKey.Key) // store securely; shown once on create
```

List API keys for an owner (master key requires the `owner` query parameter):

```go
keys, resp, err := client.ApiKeys.List(&blnkgo.ListApiKeysOptions{
    Owner: "team_acme",
})
if err != nil {
    log.Fatal(err)
}
for _, key := range keys {
    fmt.Println(key.ApiKeyID, key.Name, key.Scopes)
}
```

Revoke an API key by ID (master key requires the `owner` query parameter):

```go
resp, err := client.ApiKeys.Delete("api_key_abc123", &blnkgo.DeleteApiKeysOptions{
    Owner: "team_acme",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Revoked, status:", resp.StatusCode) // 204 No Content
```

### Hooks

Register a pre- or post-transaction webhook (master key required):

```go
hook, resp, err := client.Hooks.Create(blnkgo.CreateHookRequest{
    Name:       "Pre-transaction validation",
    URL:        "https://api.example.com/validate",
    Type:       blnkgo.HookTypePreTransaction,
    Active:     true,
    Timeout:    30,
    RetryCount: 3,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Hook ID:", hook.ID)
```

Update an existing hook by ID:

```go
updated, resp, err := client.Hooks.Update(hook.ID, blnkgo.UpdateHookRequest{
    Name:       "Pre-transaction validation (updated)",
    URL:        "https://api.example.com/validate-v2",
    Type:       blnkgo.HookTypePreTransaction,
    Active:     false,
    Timeout:    45,
    RetryCount: 5,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("Updated hook:", updated.Name)
```

View a hook by ID:

```go
hook, resp, err := client.Hooks.Get("hk_test_123")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Hook:", hook.Name, hook.Active)
```

List hooks, optionally filtered by type:

```go
hooks, resp, err := client.Hooks.List(&blnkgo.ListHooksOptions{
    Type: blnkgo.HookTypePreTransaction,
})
if err != nil {
    log.Fatal(err)
}
for _, hook := range hooks {
    fmt.Println(hook.ID, hook.Name, hook.Type)
}
```

Delete a hook by ID:

```go
deleted, resp, err := client.Hooks.Delete("hook_test_123")
if err != nil {
    log.Fatal(err)
}
fmt.Println(deleted.Message) // hook deleted successfully
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

Start a Typesense reindex and poll progress:

```go
started, resp, err := client.Search.StartReindex(&blnkgo.StartReindexRequest{
    BatchSize: intPtr(1000),
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(started.Message, started.Progress.Status)

progress, resp, err := client.Search.GetReindexStatus()
if err != nil {
    log.Fatal(err)
}
fmt.Println(progress.Status, progress.Phase, progress.ProcessedRecords, progress.TotalRecords)
```

### Error Handling

Core 0.15.0+ returns structured errors with an `error_detail` object. When a service method returns an error, use `errors.As` to read the stable machine code — do not branch on message text:

```go
_, resp, err := client.Transaction.Get("txn_missing")
if err != nil {
    var apiErr *blnkgo.ApiErrorResponse
    if errors.As(err, &apiErr) && apiErr.ErrorDetail != nil {
        switch apiErr.ErrorDetail.Code {
        case "TXN_NOT_FOUND":
            // handle missing transaction
        case "GEN_CONFLICT":
            // handle conflict
        default:
            fmt.Printf("API error %s: %s\n", apiErr.ErrorDetail.Code, apiErr.ErrorDetail.Message)
        }
    }
    return
}
```

Legacy responses with only a flat `"error"` string are mapped to `ErrorDetail.Code == "UNKNOWN"`. The raw response body remains available on `ApiErrorResponse.Body` for backward compatibility.

See [Blnk error codes](https://docs.blnkfinance.com/advanced/error-codes) for the full catalog.

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
