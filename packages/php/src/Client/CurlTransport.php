<?php

namespace Outpipe\Client;

use Outpipe\Contracts\HttpTransport;
use Outpipe\Contracts\Response;
use Outpipe\Exceptions\TransportException;

final class CurlTransport implements HttpTransport
{
    public function send(string $method, string $url, array $headers, ?string $body, float $timeout): Response
    {
        $handle = curl_init($url);

        if ($handle === false) {
            throw new TransportException('Unable to initialize the HTTP client.');
        }

        curl_setopt_array($handle, [
            CURLOPT_CUSTOMREQUEST => $method,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HEADER => true,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_TIMEOUT_MS => (int) max(1, $timeout * 1000),
            CURLOPT_CONNECTTIMEOUT_MS => (int) max(1, min($timeout, 5) * 1000),
        ]);

        if ($body !== null) {
            curl_setopt($handle, CURLOPT_POSTFIELDS, $body);
        }

        $result = curl_exec($handle);
        $error = curl_error($handle);
        $headerSize = (int) curl_getinfo($handle, CURLINFO_HEADER_SIZE);
        $status = (int) curl_getinfo($handle, CURLINFO_RESPONSE_CODE);
        curl_close($handle);

        if ($result === false) {
            throw new TransportException($error !== '' ? $error : 'The request failed.');
        }

        $rawHeaders = substr($result, 0, $headerSize);
        $responseBody = substr($result, $headerSize);

        return new Response($status, self::parseHeaders($rawHeaders), $responseBody);
    }

    private static function parseHeaders(string $rawHeaders): array
    {
        $headers = [];

        foreach (explode("\r\n", trim($rawHeaders)) as $line) {
            if (!str_contains($line, ':')) {
                continue;
            }

            [$name, $value] = explode(':', $line, 2);
            $headers[strtolower(trim($name))] = trim($value);
        }

        return $headers;
    }
}
