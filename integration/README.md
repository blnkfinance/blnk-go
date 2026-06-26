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
go test -tags=integration -v ./integration/... -run Issue73
go test -tags=integration -v ./integration/... -run Issue36

# all integration tests
go test -tags=integration -v ./integration/...
```

## Core version notes

| Issue | Core version | Why |
|-------|--------------|-----|
| #73 search identities | 0.14.x+ | Not 0.15-only |
| #36 transaction lineage | 0.14.x+ | Not 0.15-only |
| 0.15-only features (delete identity, hooks, api-keys, etc.) | 0.15.0 | Test when we reach Go 1.3.0 issues |

If `BLNK_API_KEY` is unset, integration tests skip.

## Postman

Manual request seeds live in `manual/postman/`. Use those after integration tests pass.
