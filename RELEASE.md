# Release Notes

## v1.4.0

v1.4.0 targets **Blnk Core 0.15.3**. v1.3.0 shipped Core 0.15.0 parity; this release adds dry-run previews, General Ledger `indicator` on create, refund narration/metadata, and named Core error codes.

See [Dry-run transactions](https://docs.blnkfinance.com/transactions/dry-run) and the [Core changelog](https://docs.blnkfinance.com/changelog/blnk-core).

### Transactions

- **`DryRun`** — Optional on create, bulk create, refund, inflight update, bulk commit, and bulk void request bodies. Core returns HTTP 200 with a preview (`would_apply`, `rejection`, `balances`). Nothing is written; `reference` is not consumed.
- **Typed previews** — Use `CreateDryRun`, `CreateBulkDryRun`, `RefundDryRun`, `UpdateDryRun`, `BulkCommitInflightDryRun`, and `BulkVoidInflightDryRun`. These return `TransactionPreview` / `BulkTransactionPreview` instead of a recorded transaction. Calling the posted methods with `DryRun: true` returns an error so a preview cannot be decoded as a posted transaction.
- **`Transaction.Refund`** — Accepts `Description` and `MetaData` in addition to `SkipQueue`. Empty description inherits the original; metadata is merged onto the inherited copy.

### Balances

- **`LedgerBalance.Create`** — Optional `Indicator` (must start with `@`, no spaces) when `LedgerID` is `general_ledger_id` (`GeneralLedgerID`). Duplicate indicator + currency returns `409` / `GEN_CONFLICT`.

### Hooks

- **`Hooks.List(nil)`** — `Type` remains optional. Omitting it lists PRE and POST hooks (`GET /hooks`).

### Errors

- Exported constants `ErrorCodeTxnInvalidAmount`, `ErrorCodeGenConflict`, and `ErrorCodeTxnValidationError`. Core 0.15.3 uses `TXN_VALIDATION_ERROR` for negative amounts and for source equal to destination; duplicate GL indicators return `GEN_CONFLICT`. Branch on `ApiErrorResponse.ErrorDetail.Code`.
