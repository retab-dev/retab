To run the tests:

On the production server `https://api.retab.com`:
```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --production
```

On the local server `http://localhost:4000`:
```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --local
```

Live runs preflight `RETAB_API_BASE_URL/v1/files?limit=1` once per worker. A
missing or unreachable local server skips live tests; a reachable server that
rejects `RETAB_API_KEY` fails immediately with a credential error.

To run a single test:

```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --local -k test_extract_openai[parse-sync]
```
