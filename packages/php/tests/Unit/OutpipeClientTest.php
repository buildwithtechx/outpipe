<?php

declare(strict_types=1);

namespace Outpipe\Tests\Unit;

use Outpipe\Client\OutpipeClient;
use Outpipe\Contracts\HttpTransport;
use Outpipe\Contracts\Response;
use Outpipe\Exceptions\ApiException;
use Outpipe\Laravel\Http\WebhookSignature;
use PHPUnit\Framework\TestCase;

final class OutpipeClientTest extends TestCase
{
    public function testItSendsScopedTunnelRequests(): void
    {
        $transport = new FakeTransport([
            new Response(201, [], '{"id":"tunnel-1","status":"created"}'),
            new Response(204, [], ''),
        ]);
        $client = new OutpipeClient('https://api.outpipe.dev/', 'key-123', 3, $transport);

        $tunnel = $client->createTunnel('org-1', ['name' => 'preview', 'protocol' => 'https']);
        $client->setTunnelStatus('tunnel-1', 'active');

        self::assertSame('tunnel-1', $tunnel['id']);
        self::assertSame('/api/v1/organizations/org-1/tunnels', $transport->requests[0]['path']);
        self::assertSame('Bearer key-123', $transport->requests[0]['headers']['authorization']);
        self::assertSame('active', $transport->requests[1]['body']['status']);
    }

    public function testItAddsQueryParametersToGetRequests(): void
    {
        $transport = new FakeTransport([new Response(200, [], '{"available":true}')]);
        $client = new OutpipeClient('https://api.outpipe.dev', null, 10, $transport);

        self::assertSame(['available' => true], $client->slugAvailable('new-team'));
        self::assertSame('https://api.outpipe.dev/api/v1/organizations/slug-availability?slug=new-team', $transport->requests[0]['url']);
        self::assertArrayNotHasKey('authorization', $transport->requests[0]['headers']);
    }

    public function testItExposesApiErrors(): void
    {
        $transport = new FakeTransport([new Response(422, [], '{"message":"invalid tunnel"}')]);
        $client = new OutpipeClient('https://api.outpipe.dev', 'key-123', 10, $transport);

        $this->expectException(ApiException::class);
        $this->expectExceptionMessage('invalid tunnel');
        $client->tunnel('tunnel-1');
    }

    public function testWebhookSignaturesAcceptPrefixedAndRawValues(): void
    {
        $payload = '{"event":"tunnel.ready"}';
        $secret = 'webhook-secret';
        $signature = hash_hmac('sha256', $payload, $secret);

        self::assertTrue(WebhookSignature::verify($payload, $signature, $secret));
        self::assertTrue(WebhookSignature::verify($payload, "sha256={$signature}", $secret));
        self::assertFalse(WebhookSignature::verify($payload, 'sha256=invalid', $secret));
    }
}

final class FakeTransport implements HttpTransport
{
    public array $requests = [];

    public function __construct(private array $responses) {}

    public function send(string $method, string $url, array $headers, ?string $body, float $timeout): Response
    {
        $this->requests[] = [
            'method' => $method,
            'url' => $url,
            'path' => parse_url($url, PHP_URL_PATH) . (parse_url($url, PHP_URL_QUERY) ? '?' . parse_url($url, PHP_URL_QUERY) : ''),
            'headers' => self::headers($headers),
            'body' => $body === null ? null : json_decode($body, true),
            'timeout' => $timeout,
        ];

        return array_shift($this->responses);
    }

    private static function headers(array $headers): array
    {
        $result = [];

        foreach ($headers as $header) {
            [$name, $value] = explode(':', $header, 2);
            $result[strtolower(trim($name))] = trim($value);
        }

        return $result;
    }
}
