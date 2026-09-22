<?php

namespace Outpipe\Contracts;

final readonly class Response
{
    public function __construct(
        public int $status,
        public array $headers,
        public string $body,
    ) {}

    public function json(): array
    {
        $decoded = json_decode($this->body, true);

        return is_array($decoded) ? $decoded : [];
    }
}
