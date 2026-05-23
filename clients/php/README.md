# Retab PHP SDK

<!-- @oagen-ignore-file -->

The official PHP client for the [Retab API](https://retab.com) — structured
extraction, OCR, document classification, and workflow execution backed by
LLMs.

```bash
composer require retab/retab
```

## Quick start

```php
<?php
require 'vendor/autoload.php';

use Retab\Client;

// API key from the RETAB_API_KEY env var, or passed explicitly:
$client = new Client(apiKey: getenv('RETAB_API_KEY'));

// Extract structured data from a PDF — accept any input shape:
$response = $client->extractions()->create(
    document: __DIR__ . '/invoice.pdf', // path string
    jsonSchema: [
        'type' => 'object',
        'properties' => [
            'invoice_number' => ['type' => 'string'],
            'total' => ['type' => 'number'],
        ],
    ],
);

echo $response->id . PHP_EOL;
```

## Ergonomic document input

Every endpoint that takes a `document` (and similar `MimeData`-typed fields)
accepts any of these shapes — the SDK normalises before sending:

```php
use Retab\Resource\MimeData;

// Path string
$client->extractions()->create(document: '/path/to/invoice.pdf', ...);

// URL
$client->extractions()->create(document: 'https://example.com/invoice.pdf', ...);

// SplFileInfo
$client->extractions()->create(document: new SplFileInfo('/path/to/invoice.pdf'), ...);

// Open stream
$client->extractions()->create(document: fopen('/path/to/invoice.pdf', 'rb'), ...);

// Pre-built MimeData
$client->extractions()->create(
    document: new MimeData(filename: 'invoice.pdf', url: 'data:application/pdf;base64,...'),
    ...
);
```

## Per-call overrides

```php
use Retab\RequestOptions;

$client->extractions()->create(
    document: __DIR__ . '/invoice.pdf',
    jsonSchema: [...],
    options: new RequestOptions(
        apiKey: 'sk-override-for-this-call',
        timeout: 120,
        idempotencyKey: 'extraction-2026-05-23-001',
    ),
);
```

## Pagination

```php
foreach ($client->extractions()->list()->getIterator() as $extraction) {
    echo $extraction->id . PHP_EOL;
}

// Walk forward explicitly across pages:
$page = $client->extractions()->list();
while ($page !== null) {
    foreach ($page as $row) { /* ... */ }
    $page = $page->nextPage();
}
```

## Requirements

- PHP 8.2 or newer
- Guzzle 7

## License

MIT — see [LICENSE](LICENSE).

## Other Retab SDKs

- Python: `pip install retab` ([retab-dev/retab](https://github.com/retab-dev/retab))
- Node: `npm install @retab/node`
- Documentation: <https://docs.retab.com>
