<?php

namespace Outpipe\Contracts;

interface HttpTransport
{
    public function send(string $method, string $url, array $headers, ?string $body, float $timeout): Response;
}
