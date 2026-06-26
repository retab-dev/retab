<?php

declare(strict_types=1);

// @oagen-ignore-file
//
// Hand-maintained file. The generator emits this on first run and then
// leaves it alone; spec changes do not touch it. Mirrors the ergonomic
// MimeData input handling from the Python (`prepare_mime_document`),
// Node (`coerceMimeData`), Go (`InferMIMEData`), Rust (`From<T>` impls),
// and Ruby (`MimeData.coerce`) SDKs.

namespace Retab\Resource;

use Retab\HttpClient;

/**
 * Static helpers for building `MimeData` instances from ergonomic input.
 *
 * Lives in a separate file so the generator doesn't trample it on every
 * spec regeneration. `MimeData` itself is generated; this class is the
 * stable extension surface that resource methods route caller input
 * through before constructing a request payload.
 */
final class MimeDataCoerce
{
    /**
     * Coerce a document endpoint input into URL-backed `MimeData`.
     *
     * The document-create routes (extractions, parses, partitions, splits,
     * classifications, edits) and workflow runs accept URL-backed MimeData
     * (`{filename, url}`) only. A durable file reference — a `FileRef`
     * instance or a `{id, filename, mime_type}` array — is resolved
     * client-side into a fresh download link via `GET
     * /v1/files/{id}/download-link` (mirrors the Node/CLI SDKs). Everything
     * else flows through `coerce()` unchanged.
     *
     * @param FileRef|MimeData|\SplFileInfo|string|resource|array{id?: string, filename?: string, mime_type?: string, url?: string} $input
     */
    public static function coerceDocument(mixed $input, ?HttpClient $client = null): MimeData
    {
        return self::coerce($input, $client);
    }

    /**
     * Coerce any supported input shape into a `MimeData` instance.
     *
     * Supported inputs:
     *   - `Retab\Resource\MimeData`        (passthrough)
     *   - `Retab\Resource\FileRef`         (resolved via download-link endpoint)
     *   - `SplFileInfo`                     (reads file, base64-encodes into data URL)
     *   - `string` (URL: http://, https://, data:, gs://)  — wrapped as-is
     *   - `string` (path on disk)            — read + base64-encoded
     *   - `resource` (open stream)           — read + base64-encoded
     *   - `array{filename?: string, url: string}` — already-built wire shape
     *   - `array{id: string, filename?: string, mime_type?: string}` — file-id ref
     *
     * The optional `$client` is only consulted when `$input` is a durable
     * file reference (`FileRef` or a `{id: ...}` array); resource methods
     * thread their HTTP client in so the file id can be resolved.
     *
     * @param FileRef|MimeData|\SplFileInfo|string|resource|array{id?: string, filename?: string, mime_type?: string, url?: string} $input
     */
    public static function coerce(mixed $input, ?HttpClient $client = null): MimeData
    {
        if ($input instanceof MimeData) {
            return $input;
        }
        if ($input instanceof FileRef) {
            return self::resolveFileId($input->id, $input->filename, $client);
        }
        if ($input instanceof \SplFileInfo) {
            return self::fromFile($input->getPathname(), $input->getFilename());
        }
        if (is_resource($input)) {
            $bytes = stream_get_contents($input);
            return self::fromBytes((string) $bytes, 'document');
        }
        if (is_array($input)) {
            // Already-built wire shape: `{filename?, url}`. The `url` element is
            // typed `string` by this method's `@param` shape, so `isset()` is the
            // meaningful guard; the `MimeData` constructor's typed `string` params
            // enforce the type at runtime for callers passing untyped arrays.
            if (isset($input['url'])) {
                return new MimeData(
                    filename: is_string($input['filename'] ?? null) ? $input['filename'] : 'document',
                    url: $input['url'],
                );
            }
            // Durable file reference: `{id, filename?, mime_type?}`. As above, the
            // `id` element is typed `string`; `resolveFileId()`'s `string` param
            // enforces it at runtime.
            if (isset($input['id'])) {
                $filename = is_string($input['filename'] ?? null) ? $input['filename'] : '';
                return self::resolveFileId($input['id'], $filename, $client);
            }
            throw new \InvalidArgumentException(
                'Cannot coerce array to Retab\\Resource\\MimeData; supply '
                . 'either {filename?, url} (wire shape) or {id, filename?, mime_type?} (file reference).',
            );
        }
        if (is_string($input)) {
            if (preg_match('#^(https?|gs|data):#i', $input) === 1) {
                $parsed = parse_url($input);
                $name = isset($parsed['path']) ? basename($parsed['path']) : 'document';
                return new MimeData(filename: $name !== '' ? $name : 'document', url: $input);
            }
            if (is_file($input) && is_readable($input)) {
                return self::fromFile($input, basename($input));
            }
            // Raw payload — base64 it as text.
            return self::fromBytes($input, 'document', 'text/plain');
        }
        throw new \InvalidArgumentException(
            'Cannot coerce ' . get_debug_type($input) . ' to Retab\\Resource\\MimeData; '
            . 'supply a MimeData, FileRef, SplFileInfo, path string, URL string, stream resource, or array.',
        );
    }

    /**
     * Resolve a durable file id into URL-backed `MimeData` via the Files
     * download-link endpoint (`GET /v1/files/{id}/download-link`). The
     * document routes accept MimeData only, so a stored file id is resolved
     * into a fresh signed download URL (mirrors the Node SDK and the CLI's
     * `resolveFileIDToMIMEData`).
     *
     * Prefers the durable `mime_data` reference returned by the endpoint when
     * present, falling back to the signed `download_url`.
     */
    private static function resolveFileId(string $fileId, string $filename, ?HttpClient $client): MimeData
    {
        if ($client === null) {
            throw new \InvalidArgumentException(
                'A file-id document requires a Retab client to resolve a download link; '
                . 'call the resource method on a client instance.',
            );
        }
        $link = $client->request(
            method: 'GET',
            path: '/v1/files/' . rawurlencode($fileId) . '/download-link',
        );

        $mimeData = is_array($link['mime_data'] ?? null) ? $link['mime_data'] : null;
        $downloadUrl = is_string($link['download_url'] ?? null) ? $link['download_url'] : '';
        $url = $mimeData !== null && is_string($mimeData['url'] ?? null) && $mimeData['url'] !== ''
            ? $mimeData['url']
            : $downloadUrl;
        if ($url === '') {
            throw new \InvalidArgumentException(
                "Resolving file id '$fileId': server returned no download URL.",
            );
        }

        $resolvedName = $filename;
        if ($resolvedName === '' && $mimeData !== null && is_string($mimeData['filename'] ?? null)) {
            $resolvedName = $mimeData['filename'];
        }
        if ($resolvedName === '' && is_string($link['filename'] ?? null)) {
            $resolvedName = $link['filename'];
        }
        if ($resolvedName === '') {
            $resolvedName = 'document';
        }

        return new MimeData(filename: $resolvedName, url: $url);
    }

    private static function fromFile(string $path, string $filename): MimeData
    {
        $bytes = (string) file_get_contents($path);
        $mime = self::guessMimeType($filename)
            ?? (function_exists('mime_content_type') ? (mime_content_type($path) ?: null) : null)
            ?? 'application/octet-stream';
        return new MimeData(
            filename: $filename,
            url: 'data:' . $mime . ';base64,' . base64_encode($bytes),
        );
    }

    private static function fromBytes(string $bytes, string $filename, ?string $mimeOverride = null): MimeData
    {
        $mime = $mimeOverride ?? self::guessMimeType($filename) ?? 'application/octet-stream';
        return new MimeData(
            filename: $filename,
            url: 'data:' . $mime . ';base64,' . base64_encode($bytes),
        );
    }

    private const EXTENSION_MIME_MAP = [
        'pdf' => 'application/pdf',
        'png' => 'image/png',
        'jpg' => 'image/jpeg',
        'jpeg' => 'image/jpeg',
        'gif' => 'image/gif',
        'webp' => 'image/webp',
        'txt' => 'text/plain',
        'csv' => 'text/csv',
        'json' => 'application/json',
        'xml' => 'application/xml',
        'html' => 'text/html',
        'htm' => 'text/html',
        'md' => 'text/markdown',
        'docx' => 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        'xlsx' => 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'pptx' => 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    ];

    private static function guessMimeType(string $filename): ?string
    {
        $ext = strtolower(pathinfo($filename, PATHINFO_EXTENSION));
        return self::EXTENSION_MIME_MAP[$ext] ?? null;
    }
}
