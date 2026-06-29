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
go test -tags=integration -v ./integration/... -run Issue69
go test -tags=integration -v ./integration/... -run Issue68
go test -tags=integration -v ./integration/... -run Issue71
go test -tags=integration -v ./integration/... -run Issue40
go test -tags=integration -v ./integration/... -run Issue73
go test -tags=integration -v ./integration/... -run Issue72
go test -tags=integration -v ./integration/... -run Issue55
go test -tags=integration -v ./integration/... -run Issue41
go test -tags=integration -v ./integration/... -run Issue42
go test -tags=integration -v ./integration/... -run Issue43
go test -tags=integration -v ./integration/... -run Issue44
go test -tags=integration -v ./integration/... -run Issue36
go test -tags=integration -v ./integration/... -run Issue61
go test -tags=integration -v ./integration/... -run Issue58
go test -tags=integration -v ./integration/... -run Issue59

# all integration tests
go test -tags=integration -v ./integration/... 
```

## Issue #40 — RecoverQueue workflow

1. Start Core 0.14.5: `cd ../blnk && docker compose up -d`
2. Export API key: `export BLNK_API_KEY=blnk-local-dev-secret-change-me`
3. Unit tests: `go test ./... -run RecoverQueue`
4. Integration: `go test -tags=integration -v ./integration/... -run Issue40`

## Core version notes

| Issue | Core version | Why |
|-------|--------------|-----|
| #40 recover queue | 0.14.x+ | Not 0.15-only |
| #70 refund skip_queue | 0.14.x+ | Optional body on POST /refund-transaction/{id} |
| #71 precise_distribution | 0.14.x+ | Split legs with precise_distribution on POST /transactions |
| #86 update skip_queue | 0.14.x+ | skip_queue on inflight commit/void (Update + bulk) |
| #69 balance from_source | 0.14.x+ | GET /balances/{id}?from_source=true |
| #68 balance lineage create | 0.14.x+ | track_fund_lineage + allocation_strategy on POST /balances |
| #73 search identities | 0.14.x+ | Not 0.15-only |
| #72 identity optional id | 0.14.x+ | Caller-supplied identity_id on POST /identities |
| #55 identity filter | 0.14.x+ | POST /identities/filter |
| #41 instant reconciliation | 0.14.x+ | POST /reconciliation/start-instant |
| #42 get reconciliation | 0.14.x+ | GET /reconciliation/{reconciliation_id} |
| #43 update matching rule | 0.14.x+ | PUT /reconciliation/matching-rules/{rule_id} |
| #44 delete matching rule | 0.14.x+ | DELETE /reconciliation/matching-rules/{rule_id} |
| #36 transaction lineage | 0.14.x+ | Not 0.15-only |
| #61 health check | 0.14.x+ | GET /health; auth not required |
| #58 create api key | 0.14.x+ | POST /api-keys; master or api-keys:write key |
| #59 list api keys | 0.14.x+ | GET /api-keys?owner=; master requires owner |
| #60 delete api key | 0.14.x+ | DELETE /api-keys/{id}?owner=; master requires owner |
| #50 create hook | 0.8.4+ | POST /hooks; master key required |
| #51 update hook | 0.8.4+ | PUT /hooks/{id}; master key required |
| #52 get hook | 0.8.4+ | GET /hooks/{id}; master key required |
| #53 list hooks | 0.8.4+ | GET /hooks?type=; master key required |
| #54 delete hook | 0.8.4+ | DELETE /hooks/{id}; master key required |
| #56 start reindex | 0.13.2+ | POST /search/reindex |
| #57 get reindex status | 0.13.2+ | GET /search/reindex |
| #45 tokenize identity field | 0.13.2+ | POST /identities/{id}/tokenize/{field}; tokenization must be enabled |
| #46 tokenize identity | 0.13.2+ | POST /identities/{id}/tokenize; tokenization must be enabled |
| #47 get tokenized fields | 0.13.2+ | GET /identities/{id}/tokenized-fields; tokenization must be enabled |
| #48 detokenize identity field | 0.13.2+ | GET /identities/{id}/detokenize/{field}; tokenization must be enabled |
| #49 detokenize identity | 0.13.2+ | POST /identities/{id}/detokenize; tokenization must be enabled |
| #117 delete identity | 0.15.0+ | DELETE /identities/{id} |
| #118 delete balance monitor | 0.15.0+ | DELETE /balance-monitors/{id} |
| #119 reconciliation run response | 0.15.0+ | POST /reconciliation/start returns reconciliation_id only |

If `BLNK_API_KEY` is unset, integration tests skip.
