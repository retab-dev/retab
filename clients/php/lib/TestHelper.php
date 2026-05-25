<?php

declare(strict_types=1);

// @oagen-ignore-file

namespace Retab;

use GuzzleHttp\Handler\MockHandler;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Middleware;
use GuzzleHttp\Psr7\Response;
use Psr\Http\Message\RequestInterface;

trait TestHelper
{
    /** @var list<array{request: RequestInterface}> */
    private array $recordedRequests = [];

    /**
     * @param list<array{status?: int, body?: mixed, headers?: array<string, string>}> $responses
     */
    protected function createMockClient(array $responses): Client
    {
        $mockResponses = array_map(
            static function (array $response): Response {
                $body = $response['body'] ?? [];
                return new Response(
                    $response['status'] ?? 200,
                    $response['headers'] ?? ['Content-Type' => 'application/json'],
                    is_string($body) ? $body : (string) json_encode($body),
                );
            },
            $responses,
        );

        $mock = new MockHandler($mockResponses);
        $stack = HandlerStack::create($mock);
        $stack->push(Middleware::tap(function (RequestInterface $request): void {
            $this->recordedRequests[] = ['request' => $request];
        }));

        return new Client(
            apiKey: 'test-key',
            baseUrl: 'http://localhost:4000',
            maxRetries: 0,
            handler: $stack,
        );
    }

    /**
     * @return array<string, mixed>
     */
    protected function loadFixture(string $name): array
    {
        $path = __DIR__ . '/../tests/Fixtures/' . $name . '.json';
        $contents = file_get_contents($path);
        if ($contents === false) {
            throw new \RuntimeException(sprintf('Unable to read fixture %s.', $path));
        }
        $decoded = json_decode($contents, true);
        if (!is_array($decoded)) {
            throw new \RuntimeException(sprintf('Fixture %s is not a JSON object.', $path));
        }

        /** @var array<string, mixed> $decoded */
        return $decoded;
    }

    protected function getLastRequest(): RequestInterface
    {
        $record = $this->recordedRequests[array_key_last($this->recordedRequests) ?? 0] ?? null;
        if ($record === null) {
            throw new \RuntimeException('No HTTP request was recorded.');
        }

        return $record['request'];
    }
}
