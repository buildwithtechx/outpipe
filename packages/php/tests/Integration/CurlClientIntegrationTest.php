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

    protected function setUp(): void
    {
        $this->router = tempnam(sys_get_temp_dir(), 'outpipe-router-');
        file_put_contents($this->router, <<<'PHP'
<?php
$path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
if ($_SERVER['REQUEST_METHOD'] === 'DELETE') {
    http_response_code(204);
    exit;
}
header('Content-Type: application/json');
if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    echo json_encode(['id' => 'tunnel-1', 'name' => 'preview', 'status' => 'active', 'publicUrl' => 'https://preview.outpipe.app']);
    exit;
}
if ($_SERVER['REQUEST_METHOD'] === 'GET' && str_ends_with($path, '/tunnels/tunnel-1')) {
    echo json_encode(['id' => 'tunnel-1', 'name' => 'preview', 'status' => 'active']);
    exit;
}
echo json_encode([['id' => 'tunnel-1', 'name' => 'preview', 'status' => 'active']]);
PHP);

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

        $created = $client->createTunnel('org/one', ['name' => 'preview', 'protocol' => 'http']);
        $listed = $client->tunnels('org/one');
        $inspected = $client->tunnel('tunnel-1');
        $client->revokeTunnel('tunnel-1');

        self::assertSame('tunnel-1', $created['id']);
        self::assertCount(1, $listed);
        self::assertSame('active', $inspected['status']);
    }
}
