<?php

declare(strict_types=1);

// @oagen-ignore-file
//
// Hand-maintained smoke test exercising every input shape the
// `Retab\Resource\MimeDataCoerce::coerce()` static factory must accept.
// Spec changes never regenerate this file.

namespace Retab\Tests;

use PHPUnit\Framework\TestCase;
use Retab\Resource\MimeData;
use Retab\Resource\MimeDataCoerce;

class MimeSmokeTest extends TestCase
{
    public function testCoerceSplFileInfo(): void
    {
        $info = new \SplFileInfo(__FILE__);
        $md = MimeDataCoerce::coerce($info);
        $this->assertSame('MimeSmokeTest.php', $md->filename);
        $this->assertStringStartsWith('data:', $md->url);
    }

    public function testCoerceUrlString(): void
    {
        $md = MimeDataCoerce::coerce('https://example.com/invoice.pdf');
        $this->assertSame('invoice.pdf', $md->filename);
        $this->assertSame('https://example.com/invoice.pdf', $md->url);
    }

    public function testCoerceDataUrlString(): void
    {
        $md = MimeDataCoerce::coerce('data:application/pdf;base64,XYZ');
        $this->assertSame('data:application/pdf;base64,XYZ', $md->url);
    }

    public function testCoerceGsUrlString(): void
    {
        $md = MimeDataCoerce::coerce('gs://my-bucket/path/to/file.pdf');
        $this->assertSame('file.pdf', $md->filename);
        $this->assertSame('gs://my-bucket/path/to/file.pdf', $md->url);
    }

    public function testCoercePathString(): void
    {
        $md = MimeDataCoerce::coerce(__FILE__);
        $this->assertSame('MimeSmokeTest.php', $md->filename);
        $this->assertStringStartsWith('data:', $md->url);
    }

    public function testCoerceStreamResource(): void
    {
        $stream = fopen('php://memory', 'rb+');
        $this->assertNotFalse($stream);
        fwrite($stream, 'hello world');
        rewind($stream);
        $md = MimeDataCoerce::coerce($stream);
        $this->assertSame('document', $md->filename);
        $this->assertStringStartsWith('data:application/octet-stream;base64,', $md->url);
        fclose($stream);
    }

    public function testCoercePassthrough(): void
    {
        $original = new MimeData(filename: 'x.pdf', url: 'data:application/pdf;base64,XYZ');
        $this->assertSame($original, MimeDataCoerce::coerce($original));
    }

    public function testCoerceArray(): void
    {
        $md = MimeDataCoerce::coerce(['filename' => 'x.pdf', 'url' => 'data:foo']);
        $this->assertSame('x.pdf', $md->filename);
        $this->assertSame('data:foo', $md->url);
    }

    public function testCoerceArrayWithoutFilename(): void
    {
        $md = MimeDataCoerce::coerce(['url' => 'data:bar']);
        $this->assertSame('document', $md->filename);
        $this->assertSame('data:bar', $md->url);
    }

    public function testCoerceUnsupportedThrows(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        // @phpstan-ignore-next-line — intentional bad input for the test.
        MimeDataCoerce::coerce(42);
    }
}
