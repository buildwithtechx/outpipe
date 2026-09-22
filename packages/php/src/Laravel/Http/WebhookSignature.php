<?php

declare(strict_types=1);

namespace Outpipe\Laravel\Http;

final class WebhookSignature
{
    public static function verify(string $payload, string $signature, string $secret): bool
    {
        $expected = hash_hmac('sha256', $payload, $secret);
        $provided = str_starts_with($signature, 'sha256=')
            ? substr($signature, 7)
            : $signature;

        return $provided !== '' && hash_equals($expected, $provided);
    }
}
