<?php

declare(strict_types=1);

// @oagen-ignore-file

namespace Retab\Exception;

class RetabException extends \RuntimeException
{
    /**
     * @param array<string, mixed>|null $responseBody
     */
    public function __construct(
        string $message,
        public readonly ?int $statusCode = null,
        public readonly ?array $responseBody = null,
        ?\Throwable $previous = null,
    ) {
        parent::__construct($message, $statusCode ?? 0, $previous);
    }
}
