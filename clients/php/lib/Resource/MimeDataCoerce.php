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
     * Coerce a workflow/document endpoint input while preserving durable file references.
     *
     * @param FileRef|MimeData|\SplFileInfo|string|resource|array{filename?: string, url: string} $input
     */
    public static function coerceDocument(mixed $input): FileRef|MimeData
    {
        if ($input instanceof FileRef) {
            return $input;
        }
        return self::coerce($input);
    }

    /**
     * Coerce any supported input shape into a `MimeData` instance.
     *
     * Supported inputs:
     *   - `Retab\Resource\MimeData`        (passthrough)
     *   - `SplFileInfo`                     (reads file, base64-encodes into data URL)
     *   - `string` (URL: http://, https://, data:, gs://)  — wrapped as-is
     *   - `string` (path on disk)            — read + base64-encoded
     *   - `resource` (open stream)           — read + base64-encoded
     *   - `array{filename?: string, url: string}` — already-built wire shape
     *
     * @param MimeData|\SplFileInfo|string|resource|array{filename?: string, url: string} $input
     */
    public static function coerce(mixed $input): MimeData
    {
        if ($input instanceof MimeData) {
            return $input;
        }
        if ($input instanceof \SplFileInfo) {
            return self::fromFile($input->getPathname(), $input->getFilename());
        }
        if (is_resource($input)) {
            $bytes = stream_get_contents($input);
            return self::fromBytes((string) $bytes, 'document');
        }
        if (is_array($input)) {
            return new MimeData(
                filename: $input['filename'] ?? 'document',
                url: $input['url'],
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
            . 'supply a MimeData, SplFileInfo, path string, URL string, stream resource, or array.',
        );
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
