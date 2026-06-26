# Postman — Go SDK local Core tests

Import these two files into Postman (one-time):

1. **Collection:** `go-sdk-local-core-tests.postman_collection.json`
2. **Environment:** `go-sdk-local-core.postman_environment.json`

## Run in Postman

1. Select environment **Blnk Go SDK — local Core 0.14.5**
2. Open collection **Blnk Go SDK — Local Core Tests**
3. Click **Run** (Collection Runner)
4. Run all — variables chain automatically

## Folders

| Folder | What it tests | Core version |
|--------|---------------|--------------|
| **Issue #73 — Search identities** | `POST /search/identities` | 0.14.5+ |
| **Issue #36 — Transaction GetLineage** | `GET /transactions/{id}/lineage` | 0.14.5+ |
| **Issue #40 — RecoverQueue** | `POST /transactions/recover` | 0.14.5+ |
| **Issue #70 — Refund skip_queue** | `POST /refund-transaction/{id}` | 0.14.5+ |
| **Issue #71 — precise_distribution** | `POST /transactions` split legs | 0.14.5+ |
| **Issue #86 — Update skip_queue** | `PUT /transactions/inflight/{id}` + bulk commit/void | 0.14.5+ |
| **Issue #69 — Balance from_source** | `GET /balances/{id}?from_source=true` | 0.14.5+ |
| **Issue #68 — Balance lineage create** | `POST /balances` with `track_fund_lineage` | 0.14.5+ |

## CLI (same tests, no Postman UI)

```bash
cd blnk-go
npx newman run postman/go-sdk-local-core-tests.postman_collection.json \
  -e postman/go-sdk-local-core.postman_environment.json
```

## Prerequisites

```bash
cd ../blnk && docker compose up -d
curl http://localhost:5001/health   # should return {"status":"UP"}
```

API key defaults to `blnk-local-dev-secret-change-me` from `blnk/blnk.json`.
