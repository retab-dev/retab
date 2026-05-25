<?php

declare(strict_types=1);

// @oagen-ignore-file

namespace Retab;

class RequestOptions
{
    /**
     * @param array<string, string> $headers
     * @param array<string, mixed> $query
     */
    public function __construct(
        public readonly array $headers = [],
        public readonly array $query = [],
        public readonly ?int $timeout = null,
    ) {}
}
