<?php

declare(strict_types=1);

// @oagen-ignore-file

namespace Retab;

/**
 * @template T
 * @implements \IteratorAggregate<int, T>
 */
class PaginatedResponse implements \IteratorAggregate, \Countable
{
    /**
     * @param list<T> $data
     * @param array{before?: string|null, after?: string|null} $listMetadata
     * @param (\Closure(string): self<T>)|null $fetchAfter
     */
    public function __construct(
        public readonly array $data,
        public readonly array $listMetadata,
        private readonly ?\Closure $fetchAfter = null,
    ) {}

    public function count(): int
    {
        return count($this->data);
    }

    /**
     * @return \Traversable<int, T>
     */
    public function getIterator(): \Traversable
    {
        foreach ($this->data as $item) {
            yield $item;
        }

        $after = $this->listMetadata['after'] ?? null;
        $fetchAfter = $this->fetchAfter;
        while ($after !== null && $after !== '' && $fetchAfter !== null) {
            $nextPage = $fetchAfter($after);
            foreach ($nextPage->data as $item) {
                yield $item;
            }
            $after = $nextPage->listMetadata['after'] ?? null;
        }
    }
}
