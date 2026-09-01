<?php

namespace Outpipe\Resources;

use ArrayIterator;
use Countable;
use IteratorAggregate;
use Traversable;

final class Collection implements Countable, IteratorAggregate
{
    public function __construct(private readonly array $items) {}

    public function count(): int
    {
        return count($this->items);
    }

    public function getIterator(): Traversable
    {
        return new ArrayIterator($this->items);
    }

    public function all(): array
    {
        return $this->items;
    }
}
