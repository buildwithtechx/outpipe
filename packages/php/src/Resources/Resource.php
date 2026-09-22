<?php

declare(strict_types=1);

namespace Outpipe\Resources;

abstract readonly class Resource
{
    public function __construct(protected array $attributes) {}

    public function id(): ?string
    {
        $id = $this->attributes['id'] ?? null;

        return is_string($id) ? $id : null;
    }

    public function toArray(): array
    {
        return $this->attributes;
    }

    public function __get(string $name): mixed
    {
        return $this->attributes[$name] ?? null;
    }
}
