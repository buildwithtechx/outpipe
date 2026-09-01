# Outpipe PHP SDK

The Outpipe PHP SDK provides a framework-neutral client for the public
Outpipe API and optional Laravel service-container integration.

## Installation

```sh
composer require outpipe/outpipe-php
```

The package requires PHP 8.2 and the cURL extension. Laravel applications
discover the service provider automatically.

## PHP client

```php
use Outpipe\Client\OutpipeClient;

$outpipe = new OutpipeClient(
    baseUrl: 'https://api.outpipe.dev',
    apiKey: getenv('OUTPIPE_API_KEY'),
);

$tunnel = $outpipe->createTunnel($organizationId, [
    'name' => 'checkout-preview',
    'protocol' => 'https',
    'targetHost' => '127.0.0.1',
    'targetPort' => 3000,
    'metadata' => ['application' => 'checkout'],
]);

$outpipe->setTunnelStatus($tunnel['id'], 'active');
```

The client supports account and organization management, tunnels, agents,
domains, usage, audit logs, API keys, webhooks, billing, and ownership
actions. It sends scoped credentials as `Authorization: Bearer` headers and
throws `ApiException` for non-2xx API responses.

## Laravel

Publish the configuration:

```sh
php artisan vendor:publish --tag=outpipe-config
```

Set `OUTPIPE_API_URL`, `OUTPIPE_API_KEY`, and optionally `OUTPIPE_TIMEOUT` in
the application environment. Resolve the client through dependency injection
or use the facade:

```php
use Outpipe\Laravel\Facades\Outpipe;

$tunnels = Outpipe::tunnels($organizationId);
```

The package also provides `php artisan outpipe:health`.

For Laravel webhook endpoints, verify the raw request body before dispatching
the event:

```php
use Outpipe\Laravel\Http\WebhookSignature;

$valid = WebhookSignature::verify(
    request()->getContent(),
    (string) request()->header('X-Outpipe-Signature'),
    (string) config('outpipe.webhook_secret'),
);
```

Reject the request when `$valid` is false. The verifier accepts both the raw
SHA-256 digest and the `sha256=` header format.

## Errors and testing

`OutpipeException` is the base exception. `ApiException` exposes the HTTP
status and decoded error payload; `TransportException` represents cURL or
connection failures. The transport is injectable, so applications can test
their integration without making network requests.
