<?php

namespace Outpipe\Exceptions;

final class ApiException extends OutpipeException
{
    public function __construct(
        string $message,
        public readonly int $status,
        public readonly array $payload = [],
    ) {
        parent::__construct($message, $status);
    }
}
