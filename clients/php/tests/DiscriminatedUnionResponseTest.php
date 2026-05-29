<?php

declare(strict_types=1);

// @oagen-ignore-file
//
// Hand-maintained regression for discriminated-union RESPONSE hydration.
//
// `GET /v1/workflows/artifacts/{artifact_id}` returns a `oneOf` of 11
// artifact variants keyed by the `operation` discriminator. The upstream
// `@workos/oagen` core collapses such a union to its FIRST model variant
// (`ExtractionWorkflowArtifact`), which would:
//
//   1. Drop every field unique to the other 10 variants (e.g. a split
//      artifact's `subdocuments`), and
//   2. Hard-throw inside `ExtractionWorkflowArtifact::fromArray()` when the
//      wire shape is a non-first variant, because the extraction-only
//      required fields are absent.
//
// The PHP emitter routes the response through a `match($response['operation'])`
// dispatcher that calls the correct variant's `::fromArray()`, so the round
// trip is lossless. This suite decodes a real `split` payload and asserts the
// non-first variant hydrates into `SplitWorkflowArtifact` with `subdocuments`
// intact — and that an unknown discriminator value throws rather than leaking
// a raw array.

namespace Tests;

use PHPUnit\Framework\TestCase;
use Retab\Resource\ExtractionWorkflowArtifact;
use Retab\Resource\SplitWorkflowArtifact;
use Retab\TestHelper;

class DiscriminatedUnionResponseTest extends TestCase
{
    use TestHelper;

    /**
     * A minimal-but-valid `split` artifact wire payload. Only the fields
     * `SplitWorkflowArtifact::fromArray()` requires are populated; the
     * discriminator `operation` is `split`, i.e. the SECOND union variant.
     *
     * @return array<string, mixed>
     */
    private function splitArtifactPayload(): array
    {
        return [
            'operation' => 'split',
            'id' => 'splt_123',
            'file' => [
                'id' => 'file_123',
                'filename' => 'invoice.pdf',
                'mime_type' => 'application/pdf',
            ],
            'model' => 'gpt-4o-mini',
            'subdocuments' => [
                ['name' => 'cover_letter'],
                ['name' => 'invoice_body'],
            ],
            'output' => [
                ['name' => 'cover_letter', 'pages' => [1]],
                ['name' => 'invoice_body', 'pages' => [2, 3]],
            ],
            'created_at' => '2026-05-29T12:00:00.000+00:00',
        ];
    }

    public function testSplitArtifactHydratesNonFirstVariantLosslessly(): void
    {
        $client = $this->createMockClient([
            ['status' => 200, 'body' => $this->splitArtifactPayload()],
        ]);

        // Before the emitter fix this collapsed to the first variant and
        // threw inside ExtractionWorkflowArtifact::fromArray().
        $artifact = $client->workflows()->artifacts()->get('splt_123');

        // The dispatcher must select the SECOND variant, not variants[0].
        $this->assertInstanceOf(
            SplitWorkflowArtifact::class,
            $artifact,
            'A `split` discriminator must hydrate SplitWorkflowArtifact, not the first union variant.',
        );
        $this->assertNotInstanceOf(
            ExtractionWorkflowArtifact::class,
            $artifact,
            'The union response must not collapse every variant to ExtractionWorkflowArtifact.',
        );

        // The split-only field must survive the round trip.
        $this->assertCount(2, $artifact->subdocuments);
        $this->assertSame('cover_letter', $artifact->subdocuments[0]->name);
        $this->assertSame('invoice_body', $artifact->subdocuments[1]->name);
        $this->assertSame('split', $artifact->operation);
    }

    public function testUnknownDiscriminatorThrowsRatherThanLeakingRawArray(): void
    {
        $payload = $this->splitArtifactPayload();
        $payload['operation'] = 'totally_unknown_operation';

        $client = $this->createMockClient([
            ['status' => 200, 'body' => $payload],
        ]);

        $this->expectException(\UnexpectedValueException::class);
        $client->workflows()->artifacts()->get('xxxx_123');
    }
}
