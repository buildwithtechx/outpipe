<?php

namespace Outpipe\Laravel\Console;

use Illuminate\Console\Command;
use Outpipe\Client\OutpipeClient;
use Throwable;

final class HealthCommand extends Command
{
    protected $signature = 'outpipe:health';
    protected $description = 'Check connectivity to the Outpipe API';

    public function handle(OutpipeClient $client): int
    {
        try {
            $client->health();
            $this->info('Outpipe API is reachable.');

            return self::SUCCESS;
        } catch (Throwable $exception) {
            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }
}
