# Integration tests

Requires Blnk Core running locally and an API key.

```bash
cd ../blnk && docker compose up -d
export BLNK_API_KEY=your_master_or_api_key
# optional: export BLNK_BASE_URL=http://localhost:5001/
```

## Run

```bash
# one issue
go test -tags=integration -v ./integration/... -run Issue70
go test -tags=integration -v ./integration/... -run Issue86
go test -tags=integration -v ./integration/... -run Issue68
go test -tags=integration -v ./integration/... -run Issue71
go test -tags=integration -v ./integration/... -run Issue40
go test -tags=integration -v ./integration/... -run Issue73
go test -tags=integration -v ./integration/... -run Issue36

# all integration tests
go test -tags=integration -v ./integration/... 
```

## Issue #40 — RecoverQueue workflow

1. Start Core 0.14.5: `cd ../blnk && docker compose up -d`
2. Export API key: `export BLNK_API_KEY=blnk-local-dev-secret-change-me`
3. Unit tests: `go test ./... -run RecoverQueue`
4. Integration: `go test -tags=integration -v ./integration/... -run Issue40`
5. Postman: run folder **Issue #40 — RecoverQueue** in `postman/go-sdk-local-core-tests.postman_collection.json`

## Core version notes

| Issue | Core version | Why |
|-------|--------------|-----|
| #40 recover queue | 0.14.x+ | Not 0.15-only |
| #70 refund skip_queue | 0.14.x+ | Optional body on POST /refund-transaction/{id} |
| #71 precise_distribution | 0.14.x+ | Split legs with precise_distribution on POST /transactions |
| #86 update skip_queue | 0.14.x+ | skip_queue on inflight commit/void (Update + bulk) |
| #68 balance lineage create | 0.14.x+ | track_fund_lineage + allocation_strategy on POST /balances |
| #73 search identities | 0.14.x+ | Not 0.15-only |
| #36 transaction lineage | 0.14.x+ | Not 0.15-only |
| 0.15-only features (delete identity, hooks, api-keys, etc.) | 0.15.0 | Test when we reach Go 1.3.0 issues |

If `BLNK_API_KEY` is unset, integration tests skip.

## Postman

See `postman/README.md`. Use the collection runner after integration tests pass.
