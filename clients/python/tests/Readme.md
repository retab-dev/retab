To run the tests:

On the production server `https://api.retab.dev`:
```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --production
```

On the local server `http://localhost:4000`:
```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --local
```

To run a single test:

```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --local -k test_extract_openai[parse-sync]
```