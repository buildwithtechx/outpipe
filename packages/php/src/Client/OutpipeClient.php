<?php

declare(strict_types=1);

namespace Outpipe\Client;

use Outpipe\Contracts\HttpTransport;
use Outpipe\Contracts\Response;
use Outpipe\Exceptions\ApiException;

final class OutpipeClient
{
    public function __construct(
        private readonly string $baseUrl,
        private readonly ?string $apiKey = null,
        private readonly float $timeout = 10.0,
        private readonly ?HttpTransport $transport = null,
    ) {}

    public function health(): array { return $this->request('GET', '/healthz'); }

    public function ready(): array { return $this->request('GET', '/readyz'); }

    public function account(): array { return $this->request('GET', '/api/v1/account'); }

    public function organizations(): array { return $this->request('GET', '/api/v1/organizations'); }

    public function organization(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId)); }

    public function createOrganization(string $name, string $slug): array
    {
        return $this->request('POST', '/api/v1/organizations', ['name' => $name, 'slug' => $slug]);
    }

    public function slugAvailable(string $slug): array
    {
        return $this->request('GET', '/api/v1/organizations/slug-availability', ['slug' => $slug]);
    }

    public function members(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/members'); }

    public function tunnels(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/tunnels'); }

    public function createTunnel(string $organizationId, array $tunnel): array
    {
        return $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/tunnels', $tunnel);
    }

    public function tunnel(string $tunnelId): array { return $this->request('GET', '/api/v1/tunnels/' . $this->segment($tunnelId)); }

    public function setTunnelStatus(string $tunnelId, string $status): void
    {
        $this->request('PATCH', '/api/v1/tunnels/' . $this->segment($tunnelId) . '/status', ['status' => $status]);
    }

    public function updateTunnelConfiguration(string $tunnelId, array $configuration): array
    {
        return $this->request('PATCH', '/api/v1/tunnels/' . $this->segment($tunnelId) . '/config', $configuration);
    }

    public function revokeTunnel(string $tunnelId): void
    {
        $this->request('DELETE', '/api/v1/tunnels/' . $this->segment($tunnelId));
    }

    public function agents(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/agents'); }

    public function registerAgent(string $organizationId, array $agent): array
    {
        return $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/agents', $agent);
    }

    public function revokeAgent(string $agentId): void
    {
        $this->request('DELETE', '/api/v1/agents/' . $this->segment($agentId));
    }

    public function domains(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/domains'); }

    public function createDomain(string $organizationId, array $domain): array
    {
        return $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/domains', $domain);
    }

    public function verifyDomain(string $domainId): array { return $this->request('POST', '/api/v1/domains/' . $this->segment($domainId) . '/verify'); }

    public function usageSnapshot(string $organizationId, array $query = []): array
    {
        return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/usage/snapshot', $query);
    }

    public function usageEvents(string $organizationId, array $query = []): array
    {
        return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/usage/events', $query);
    }

    public function usageRequests(string $organizationId, array $query = []): array
    {
        return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/usage/requests', $query);
    }

    public function auditLogs(string $organizationId, array $query = []): array
    {
        return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/audit-logs', $query);
    }

    public function apiKeys(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/api-keys'); }

    public function createApiKey(string $organizationId, array $key): array
    {
        return $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/api-keys', $key);
    }

    public function revokeApiKey(string $organizationId, string $keyId): void
    {
        $this->request('DELETE', '/api/v1/organizations/' . $this->segment($organizationId) . '/api-keys/' . $this->segment($keyId));
    }

    public function webhooks(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/webhooks'); }

    public function createWebhook(string $organizationId, array $webhook): array
    {
        return $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/webhooks', $webhook);
    }

    public function deleteWebhook(string $organizationId, string $webhookId): void
    {
        $this->request('DELETE', '/api/v1/organizations/' . $this->segment($organizationId) . '/webhooks/' . $this->segment($webhookId));
    }

    public function webhookDeliveries(string $organizationId, string $webhookId): array
    {
        return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/webhooks/' . $this->segment($webhookId) . '/deliveries');
    }

    public function billing(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing'); }

    public function plans(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing/plans'); }

    public function invoices(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing/invoices'); }

    public function checkout(string $organizationId, array $checkout): array
    {
        return $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing/checkout', $checkout);
    }

    public function billingPortal(string $organizationId): array { return $this->request('GET', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing/portal'); }

    public function cancelBilling(string $organizationId): void { $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing/cancel'); }

    public function resumeBilling(string $organizationId): void { $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/billing/resume'); }

    public function deleteAccount(): void { $this->request('DELETE', '/api/v1/account'); }

    public function transferOwnership(string $organizationId, string $userId): void
    {
        $this->request('POST', '/api/v1/organizations/' . $this->segment($organizationId) . '/transfer', ['userId' => $userId]);
    }

    private function request(string $method, string $path, ?array $payload = null): array
    {
        $query = $method === 'GET' && $payload !== null ? '?' . http_build_query($payload) : '';
        $url = rtrim($this->baseUrl, '/') . $path . $query;
        $headers = ['Accept: application/json', 'Content-Type: application/json'];

        if ($this->apiKey !== null && $this->apiKey !== '') {
            $headers[] = 'Authorization: Bearer ' . $this->apiKey;
        }

        $body = $method === 'GET' || $payload === null ? null : json_encode($payload, JSON_THROW_ON_ERROR);
        $response = ($this->transport ?? new CurlTransport())->send($method, $url, $headers, $body, $this->timeout);

        if ($response->status < 200 || $response->status >= 300) {
            throw new ApiException(self::message($response), $response->status, $response->json());
        }

        return $response->json();
    }

    private function segment(string $value): string
    {
        return rawurlencode($value);
    }

    private static function message(Response $response): string
    {
        $payload = $response->json();
        $message = $payload['message'] ?? $payload['error'] ?? null;

        return is_string($message) && $message !== '' ? $message : "Outpipe API request failed with status {$response->status}.";
    }
}
