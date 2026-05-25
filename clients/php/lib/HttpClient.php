<?php

declare(strict_types=1);

// @oagen-ignore-file

namespace Retab;

use GuzzleHttp\Client as GuzzleClient;
use GuzzleHttp\Exception\GuzzleException;
use GuzzleHttp\HandlerStack;
use Retab\Exception\RetabException;

class HttpClient
{
    private GuzzleClient $client;

    public function __construct(
        private readonly string $apiKey,
        private readonly ?string $clientId,
        string $baseUrl,
        private readonly int $timeout = 60,
        private readonly int $maxRetries = 3,
        ?HandlerStack $handler = null,
        private readonly ?string $userAgent = null,
    ) {
        $config = [
            'base_uri' => rtrim($baseUrl, '/') . '/',
            'timeout' => $timeout,
        ];
        if ($handler !== null) {
            $config['handler'] = $handler;
        }
        $this->client = new GuzzleClient($config);
    }

    /**
     * @param array<string, mixed> $query
     * @param array<string, mixed>|null $body
     * @return array<string, mixed>
     */
    public function request(
        string $method,
        string $path,
        ?array $body = null,
        array $query = [],
        ?RequestOptions $options = null,
    ): array {
        $requestOptions = $this->requestOptions($query, $options);
        if ($body !== null) {
            $requestOptions['json'] = $body;
        }

        $attempt = 0;
        while (true) {
            try {
                $response = $this->client->request($method, ltrim($path, '/'), $requestOptions);
                $statusCode = $response->getStatusCode();
                $rawBody = (string) $response->getBody();
                if ($statusCode === 204 || $rawBody === '') {
                    return [];
                }

                $decoded = json_decode($rawBody, true);
                if (!is_array($decoded)) {
                    throw new RetabException('Retab API returned an invalid JSON response.', $statusCode);
                }

                /** @var array<string, mixed> $decoded */
                return $decoded;
            } catch (GuzzleException $error) {
                if ($attempt >= $this->maxRetries) {
                    throw new RetabException($error->getMessage(), null, null, $error);
                }
                $attempt++;
            }
        }
    }

    /**
     * @template T
     * @param array<string, mixed> $query
     * @param class-string<T> $modelClass
     * @return PaginatedResponse<T>
     */
    public function requestPage(
        string $method,
        string $path,
        array $query,
        string $modelClass,
        ?RequestOptions $options = null,
    ): PaginatedResponse {
        $response = $this->request(
            method: $method,
            path: $path,
            query: $query,
            options: $options,
        );

        $data = [];
        foreach ($response['data'] ?? [] as $item) {
            if (is_array($item) && method_exists($modelClass, 'fromArray')) {
                $data[] = $modelClass::fromArray($item);
            }
        }

        $metadata = is_array($response['list_metadata'] ?? null) ? $response['list_metadata'] : [];
        /** @var array{before?: string|null, after?: string|null} $metadata */

        return new PaginatedResponse(
            data: $data,
            listMetadata: $metadata,
            fetchAfter: function (string $after) use ($method, $path, $query, $modelClass, $options): PaginatedResponse {
                $nextQuery = $query;
                unset($nextQuery['before']);
                $nextQuery['after'] = $after;
                return $this->requestPage($method, $path, $nextQuery, $modelClass, $options);
            },
        );
    }

    /**
     * @param array<string, mixed> $query
     * @return array<string, mixed>
     */
    private function requestOptions(array $query, ?RequestOptions $options): array
    {
        $mergedQuery = array_filter(
            array_merge($query, $options !== null ? $options->query : []),
            static fn($value) => $value !== null,
        );

        $headers = array_filter([
            'Authorization' => $this->apiKey !== '' ? 'Bearer ' . $this->apiKey : null,
            'X-Retab-Client-Id' => $this->clientId,
            'User-Agent' => $this->userAgent,
            'Accept' => 'application/json',
        ], static fn($value) => $value !== null);

        return [
            'headers' => array_merge($headers, $options !== null ? $options->headers : []),
            'query' => $mergedQuery,
            'timeout' => $options !== null && $options->timeout !== null ? $options->timeout : $this->timeout,
        ];
    }
}
