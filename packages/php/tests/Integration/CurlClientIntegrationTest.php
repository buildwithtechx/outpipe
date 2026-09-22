<?php

declare(strict_types=1);

namespace OutpipeTests\Integration;

use Outpipe\Client\OutpipeClient;
use PHPUnit\Framework\TestCase;

final class CurlClientIntegrationTest extends TestCase
{
    private $process;
    private string $baseUrl;
    private string $router;
    private array $contract;

    protected function setUp(): void
    {
        $contractPath = dirname(__DIR__, 4) . '/protocol/fixtures/http_tunnel_contract.json';
        $this->contract = json_decode(file_get_contents($contractPath), true, 512, JSON_THROW_ON_ERROR);
        $this->router = tempnam(sys_get_temp_dir(), 'outpipe-router-');
        $router = <<<'PHP'
<?php
$path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
$headers = array_change_key_case(getallheaders());
if (($headers['authorization'] ?? '') !== 'Bearer integration-key') {
    http_response_code(401);
    echo json_encode(['message' => 'authentication required']);
    exit;
}
if ($path === '/api/v1/organizations/error%2F401/tunnels') {
    http_response_code(401);
    echo json_encode(['message' => 'authentication required']);
    exit;
}
if ($_SERVER['REQUEST_METHOD'] === 'POST' && $path === '/api/v1/organizations/org%2Fone/tunnels') {
    $payload = json_decode(file_get_contents('php://input'), true);
    if ($payload !== ['name' => 'preview', 'protocol' => 'http', 'targetHost' => '127.0.0.1', 'targetPort' => 3000]) {
        http_response_code(422);
        echo json_encode(['message' => 'invalid tunnel']);
        exit;
    }
    http_response_code(201);
    header('Content-Type: application/json');
    echo json_encode(['id' => 'tunnel-1', 'name' => 'preview', 'status' => 'active', 'publicHostname' => 'preview.outpipe.app']);
    exit;
}
if ($_SERVER['REQUEST_METHOD'] === 'GET' && $path === '/api/v1/organizations/org%2Fone/tunnels') {
    header('Content-Type: application/json');
    echo json_encode([['id' => 'tunnel-1', 'name' => 'preview', 'status' => 'active', 'publicHostname' => 'preview.outpipe.app']]);
    exit;
}
if ($_SERVER['REQUEST_METHOD'] === 'GET' && $path === '/api/v1/tunnels/tunnel-1') {
    header('Content-Type: application/json');
    echo json_encode(['id' => 'tunnel-1', 'name' => 'preview', 'status' => 'active', 'publicHostname' => 'preview.outpipe.app']);
    exit;
}
if ($_SERVER['REQUEST_METHOD'] === 'DELETE' && $path === '/api/v1/tunnels/tunnel-1') {
    http_response_code(204);
    exit;
}
http_response_code(404);
echo json_encode(['message' => 'not found']);
PHP;
        file_put_contents($this->router, $router);

        $port = random_int(18_000, 28_000);
        $command = sprintf(
            'php -S 127.0.0.1:%d %s',
            $port,
            escapeshellarg($this->router),
        );
        $this->process = proc_open($command, [
            0 => ['pipe', 'r'],
            1 => ['pipe', 'w'],
            2 => ['pipe', 'w'],
        ], $pipes);
        self::assertIsResource($this->process);
        fclose($pipes[0]);
        stream_set_blocking($pipes[1], false);
        stream_set_blocking($pipes[2], false);

        $output = '';
        for ($attempt = 0; $attempt < 50; $attempt++) {
            $output = stream_get_contents($pipes[1]) . stream_get_contents($pipes[2]);
            $socket = @fsockopen('127.0.0.1', $port, $errorCode, $errorMessage, 0.1);
            if (is_resource($socket)) {
                fclose($socket);
                $this->baseUrl = 'http://127.0.0.1:' . $port;
                fclose($pipes[1]);
                fclose($pipes[2]);
                return;
            }
            usleep(20_000);
        }

        $this->fail('Unable to start PHP HTTP server: ' . $output);
    }

    protected function tearDown(): void
    {
        if (is_resource($this->process)) {
            proc_terminate($this->process);
            proc_close($this->process);
        }
        if (isset($this->router) && file_exists($this->router)) {
            unlink($this->router);
        }
    }

    public function testItCompletesTunnelLifecycleThroughCurlTransport(): void
    {
        $client = new OutpipeClient($this->baseUrl, 'integration-key');

        $created = $client->createTunnel('org/one', $this->contract['routes']['create']['request']);
        $listed = $client->tunnels('org/one');
        $inspected = $client->tunnel('tunnel-1');
        $client->revokeTunnel('tunnel-1');

        self::assertSame('tunnel-1', $created['id']);
        self::assertSame($this->contract['tunnel']['publicHostname'], $created['publicHostname']);
        self::assertCount(1, $listed);
        self::assertSame('active', $inspected['status']);

        try {
            $client->tunnels('error/401');
            self::fail('Expected an API exception.');
        } catch (\Outpipe\Exceptions\ApiException $exception) {
            self::assertSame(401, $exception->status);
        }
    }
}
