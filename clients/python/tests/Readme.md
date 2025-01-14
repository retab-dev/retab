To run the tests:

On the production server `https://api.uiform.com`:
```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --production
```

On the local server `http://localhost:4000`:
```bash
pytest -n auto --dist=load --asyncio-mode=strict -W error -s -v --local
```

