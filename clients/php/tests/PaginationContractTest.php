<?php

declare(strict_types=1);

// @oagen-ignore-file
//
// Hand-maintained pagination-contract regression. Mirrors the equivalent
// suites in the Python, Node, and Go SDKs
// (.notes/blueprints/sdk-pagination-contract.md).
//
// For every Retab\Service\* class that exposes a `list(...)` method
// returning `Retab\PaginatedResponse`, this suite:
//
//   1. Mocks the HTTP transport with TWO canned responses — a first page
//      that advertises a non-null `after` cursor and a second page that
//      closes it out.
//   2. Calls `list(...)` and drains the page via `iterator_to_array($page)`.
//   3. Asserts that the iterator-driven drain triggered EXACTLY two HTTP
//      requests — the first with no `after` query param, the second with
//      `after=cursor-2`. If a `list` method ever stops delegating through
//      `HttpClient::requestPage`, the fetch closure is dropped, only one
//      request fires, and this test names the offender.
//
// A second discovery test scans `lib/Service/` via glob() and asserts
// every service with a `public function list(...)` method appears in the
// registry below — so a freshly-generated service that forgets to register
// itself fails CI instead of silently bypassing the contract.

namespace Tests;

use PHPUnit\Framework\TestCase;
use Retab\Client;
use Retab\PaginatedResponse;
use Retab\TestHelper;

class PaginationContractTest extends TestCase
{
    use TestHelper;

    /**
     * Services that legitimately bypass the central pagination contract.
     * Empty for PHP — when populated, every entry needs a comment linking
     * to the matching exception in
     * .notes/blueprints/sdk-pagination-contract.md.
     *
     * @var array<int, string>
     */
    private const KNOWN_BYPASS = [];

    /**
     * Each entry pins one service's `list(...)` invocation. Keys:
     *
     *   - serviceClass:   FQCN under `lib/Service/`
     *   - clientAccessor: method name on Retab\Client that returns it
     *   - listFixture:    fixture filename (without `.json`) — the existing
     *                     per-service generated test already proves the
     *                     fixture deserialises cleanly via `fromArray`, so
     *                     reusing it here keeps coverage honest.
     *   - invoke:         closure that calls `->list(...)` on the resolved
     *                     service object. Required positional args (run id,
     *                     workflow id, etc.) are supplied with test strings.
     *
     * @return array<string, array{
     *     serviceClass: class-string,
     *     clientAccessor: string,
     *     listFixture: string,
     *     invoke: \Closure(Client): PaginatedResponse,
     * }>
     */
    private function registry(): array
    {
        return [
            'Classifications' => [
                'serviceClass' => \Retab\Service\Classifications::class,
                'clientAccessor' => 'classifications',
                'listFixture' => 'list_classification',
                'invoke' => static fn(Client $c) => $c->classifications()->list(),
            ],
            'EditTemplates' => [
                'serviceClass' => \Retab\Service\EditTemplates::class,
                'clientAccessor' => 'editTemplates',
                'listFixture' => 'list_edit_template',
                'invoke' => static fn(Client $c) => $c->editTemplates()->list(),
            ],
            'Edits' => [
                'serviceClass' => \Retab\Service\Edits::class,
                'clientAccessor' => 'edits',
                'listFixture' => 'list_edit',
                'invoke' => static fn(Client $c) => $c->edits()->list(),
            ],
            'ExperimentRunResults' => [
                'serviceClass' => \Retab\Service\ExperimentRunResults::class,
                'clientAccessor' => 'experimentRunResults',
                'listFixture' => 'list_experiment_result',
                'invoke' => static fn(Client $c) => $c->experimentRunResults()->list('test_run_id'),
            ],
            'ExperimentRuns' => [
                'serviceClass' => \Retab\Service\ExperimentRuns::class,
                'clientAccessor' => 'experimentRuns',
                'listFixture' => 'list_experiment_run',
                'invoke' => static fn(Client $c) => $c->experimentRuns()->list(),
            ],
            'Extractions' => [
                'serviceClass' => \Retab\Service\Extractions::class,
                'clientAccessor' => 'extractions',
                'listFixture' => 'list_extraction',
                'invoke' => static fn(Client $c) => $c->extractions()->list(),
            ],
            'Files' => [
                'serviceClass' => \Retab\Service\Files::class,
                'clientAccessor' => 'files',
                'listFixture' => 'list_file',
                'invoke' => static fn(Client $c) => $c->files()->list(),
            ],
            'Jobs' => [
                'serviceClass' => \Retab\Service\Jobs::class,
                'clientAccessor' => 'jobs',
                'listFixture' => 'list_job',
                'invoke' => static fn(Client $c) => $c->jobs()->list(),
            ],
            'Parses' => [
                'serviceClass' => \Retab\Service\Parses::class,
                'clientAccessor' => 'parses',
                'listFixture' => 'list_parse',
                'invoke' => static fn(Client $c) => $c->parses()->list(),
            ],
            'Partitions' => [
                'serviceClass' => \Retab\Service\Partitions::class,
                'clientAccessor' => 'partitions',
                'listFixture' => 'list_partition',
                'invoke' => static fn(Client $c) => $c->partitions()->list(),
            ],
            'Splits' => [
                'serviceClass' => \Retab\Service\Splits::class,
                'clientAccessor' => 'splits',
                'listFixture' => 'list_split',
                'invoke' => static fn(Client $c) => $c->splits()->list(),
            ],
            'WorkflowArtifacts' => [
                'serviceClass' => \Retab\Service\WorkflowArtifacts::class,
                'clientAccessor' => 'workflowArtifacts',
                'listFixture' => 'list_workflow_artifact',
                'invoke' => static fn(Client $c) => $c->workflowArtifacts()->list(),
            ],
            'WorkflowBlockExecutions' => [
                'serviceClass' => \Retab\Service\WorkflowBlockExecutions::class,
                'clientAccessor' => 'workflowBlockExecutions',
                'listFixture' => 'list_stored_block_execution',
                'invoke' => static fn(Client $c) => $c->workflowBlockExecutions()->list('test_run_id', 'test_block_id'),
            ],
            'WorkflowBlocks' => [
                'serviceClass' => \Retab\Service\WorkflowBlocks::class,
                'clientAccessor' => 'workflowBlocks',
                'listFixture' => 'list_workflow_block',
                'invoke' => static fn(Client $c) => $c->workflowBlocks()->list('test_workflow_id'),
            ],
            'WorkflowEdges' => [
                'serviceClass' => \Retab\Service\WorkflowEdges::class,
                'clientAccessor' => 'workflowEdges',
                'listFixture' => 'list_workflow_edge_doc',
                'invoke' => static fn(Client $c) => $c->workflowEdges()->list('test_workflow_id'),
            ],
            'WorkflowExperiments' => [
                'serviceClass' => \Retab\Service\WorkflowExperiments::class,
                'clientAccessor' => 'workflowExperiments',
                'listFixture' => 'list_workflow_experiment',
                'invoke' => static fn(Client $c) => $c->workflowExperiments()->list('test_workflow_id'),
            ],
            'WorkflowReviewVersions' => [
                'serviceClass' => \Retab\Service\WorkflowReviewVersions::class,
                'clientAccessor' => 'workflowReviewVersions',
                'listFixture' => 'list_review_version',
                'invoke' => static fn(Client $c) => $c->workflowReviewVersions()->list('test_review_id'),
            ],
            'WorkflowReviews' => [
                'serviceClass' => \Retab\Service\WorkflowReviews::class,
                'clientAccessor' => 'workflowReviews',
                'listFixture' => 'list_review',
                'invoke' => static fn(Client $c) => $c->workflowReviews()->list(),
            ],
            'WorkflowRuns' => [
                'serviceClass' => \Retab\Service\WorkflowRuns::class,
                'clientAccessor' => 'workflowRuns',
                'listFixture' => 'list_workflow_run',
                'invoke' => static fn(Client $c) => $c->workflowRuns()->list(),
            ],
            'WorkflowSteps' => [
                'serviceClass' => \Retab\Service\WorkflowSteps::class,
                'clientAccessor' => 'workflowSteps',
                'listFixture' => 'list_workflow_run_step',
                'invoke' => static fn(Client $c) => $c->workflowSteps()->list(),
            ],
            'WorkflowTestRunResults' => [
                'serviceClass' => \Retab\Service\WorkflowTestRunResults::class,
                'clientAccessor' => 'workflowTestRunResults',
                'listFixture' => 'list_workflow_test_result',
                'invoke' => static fn(Client $c) => $c->workflowTestRunResults()->list('test_run_id'),
            ],
            'WorkflowTestRuns' => [
                'serviceClass' => \Retab\Service\WorkflowTestRuns::class,
                'clientAccessor' => 'workflowTestRuns',
                'listFixture' => 'list_workflow_test_run',
                'invoke' => static fn(Client $c) => $c->workflowTestRuns()->list(),
            ],
            'WorkflowTests' => [
                'serviceClass' => \Retab\Service\WorkflowTests::class,
                'clientAccessor' => 'workflowTests',
                'listFixture' => 'list_workflow_test',
                'invoke' => static fn(Client $c) => $c->workflowTests()->list('test_workflow_id'),
            ],
            'Workflows' => [
                'serviceClass' => \Retab\Service\Workflows::class,
                'clientAccessor' => 'workflows',
                'listFixture' => 'list_workflow',
                'invoke' => static fn(Client $c) => $c->workflows()->list(),
            ],
        ];
    }

    /**
     * @return array<int, array{string, array{
     *     serviceClass: class-string,
     *     clientAccessor: string,
     *     listFixture: string,
     *     invoke: \Closure(Client): PaginatedResponse,
     * }}>
     */
    public static function registryProvider(): array
    {
        $self = new self('placeholder');
        $rows = [];
        foreach ($self->registry() as $name => $entry) {
            $rows[] = [$name, $entry];
        }
        return $rows;
    }

    /**
     * Auto-pagination must walk every page when the caller drains the
     * iterator. The HTTP layer is mocked with TWO pages and we assert
     * both got hit.
     *
     * @param array{
     *     serviceClass: class-string,
     *     clientAccessor: string,
     *     listFixture: string,
     *     invoke: \Closure(Client): PaginatedResponse,
     * } $entry
     */
    #[\PHPUnit\Framework\Attributes\DataProvider('registryProvider')]
    public function testIteratorWalksAllPages(string $name, array $entry): void
    {
        $fixture = $this->loadFixture($entry['listFixture']);
        $items = is_array($fixture['data'] ?? null) ? $fixture['data'] : [];

        // Synthesise page 1 with a non-null `after` cursor so the iterator
        // is obligated to fetch a follow-up page, and page 2 that closes
        // the cursor out.
        $page1 = $fixture;
        $page1['data'] = $items === [] ? [] : [$items[0]];
        $page1['list_metadata'] = ['before' => null, 'after' => 'cursor-2'];

        $page2 = $fixture;
        $page2['data'] = $items === [] ? [] : [$items[0]];
        $page2['list_metadata'] = ['before' => null, 'after' => null];

        $client = $this->createMockClient([
            ['status' => 200, 'body' => $page1],
            ['status' => 200, 'body' => $page2],
        ]);

        $result = ($entry['invoke'])($client);
        $this->assertInstanceOf(
            PaginatedResponse::class,
            $result,
            sprintf('%s::list() must return Retab\\PaginatedResponse, got %s', $name, get_debug_type($result)),
        );

        // Drain the iterator end-to-end — this is what forces the second
        // HTTP call when the closure is wired correctly.
        $drained = iterator_to_array($result, preserve_keys: false);

        // Both fixture pages emit one item, so a fully-walked iterator
        // produces two items.
        $expected = count($page1['data']) + count($page2['data']);
        $this->assertCount(
            $expected,
            $drained,
            sprintf(
                '%s: iterator yielded %d items but mock served two pages of %d total. '
                . 'Likely cause: list() constructed PaginatedResponse without the fetch closure '
                . '(bypassed HttpClient::requestPage), so auto-pagination silently stopped after page 1.',
                $name,
                count($drained),
                $expected,
            ),
        );

        // Confirm the HTTP layer was hit exactly twice — once for the
        // initial page, once for the cursor follow-up.
        $this->assertCount(
            2,
            $this->recordedRequests(),
            sprintf('%s: expected 2 HTTP requests after draining iterator, got %d.', $name, count($this->recordedRequests())),
        );

        // First request must omit `after` (or have an empty value);
        // second must carry `after=cursor-2`.
        $firstQuery = $this->parseQuery(0);
        $secondQuery = $this->parseQuery(1);

        $this->assertTrue(
            !isset($firstQuery['after']) || $firstQuery['after'] === '',
            sprintf('%s: first request should not carry an `after` cursor, got %s', $name, $firstQuery['after'] ?? '(unset)'),
        );
        $this->assertSame(
            'cursor-2',
            $secondQuery['after'] ?? null,
            sprintf('%s: second request must carry `after=cursor-2` from the previous page\'s list_metadata.', $name),
        );
        // Per the cross-language contract, the closure must drop `before`
        // when walking forward — `before` and `after` are mutually
        // exclusive on every Retab list route.
        $this->assertArrayNotHasKey(
            'before',
            $secondQuery,
            sprintf('%s: follow-up request must drop the `before` cursor when paging forward.', $name),
        );
    }

    /**
     * Glob `lib/Service/` and verify every `*.php` file with a
     * `public function list(...)` method whose return type is
     * `PaginatedResponse` shows up in the registry. A newly-generated
     * service that forgets to register itself here fails CI.
     */
    public function testRegistryCoversEveryListMethod(): void
    {
        $servicesDir = __DIR__ . '/../lib/Service';
        $files = glob($servicesDir . '/*.php');
        $this->assertNotFalse($files, 'glob() failed to read lib/Service/');

        $expected = [];
        foreach ($files as $file) {
            $svc = basename($file, '.php');
            if (in_array($svc, self::KNOWN_BYPASS, true)) {
                continue;
            }
            $contents = (string) file_get_contents($file);
            // Match a `public function list(...)` whose return type
            // declaration mentions `PaginatedResponse`. Multi-line
            // signature, so use the `s` modifier.
            if (preg_match('/public\s+function\s+list\s*\([^)]*\)\s*:\s*\\\\?Retab\\\\PaginatedResponse/s', $contents) === 1) {
                $expected[] = $svc;
            }
        }
        sort($expected);

        $registered = array_keys($this->registry());
        sort($registered);

        $missing = array_diff($expected, $registered);
        $extra = array_diff($registered, $expected);

        $this->assertSame(
            [],
            array_values($missing),
            'Services with a list()->PaginatedResponse method that are not in the pagination-contract registry: '
            . implode(', ', $missing)
            . '. Add an entry to PaginationContractTest::registry() so the contract covers it.',
        );
        $this->assertSame(
            [],
            array_values($extra),
            'Services in the registry that no longer expose list()->PaginatedResponse: '
            . implode(', ', $extra)
            . '. Drop them from PaginationContractTest::registry().',
        );
    }

    /** @return list<array{request: \Psr\Http\Message\RequestInterface}> */
    private function recordedRequests(): array
    {
        // `TestHelper::$recordedRequests` is private — reach into it via
        // reflection. `setAccessible()` was deprecated in PHP 8.5 (no-op
        // since 8.1) so we skip it.
        $prop = new \ReflectionProperty($this, 'recordedRequests');
        /** @var list<array{request: \Psr\Http\Message\RequestInterface}> $value */
        $value = $prop->getValue($this);
        return $value;
    }

    /**
     * @return array<string, string>
     */
    private function parseQuery(int $index): array
    {
        $recorded = $this->recordedRequests();
        if (!isset($recorded[$index])) {
            $this->fail(sprintf('No recorded request at index %d (have %d).', $index, count($recorded)));
        }
        $request = $recorded[$index]['request'];
        parse_str($request->getUri()->getQuery(), $query);
        /** @var array<string, string> $query */
        return $query;
    }
}
