<?php

namespace Outpipe\Laravel;

use Illuminate\Support\ServiceProvider;
use Illuminate\Contracts\Config\Repository;
use Illuminate\Contracts\Foundation\Application;
use Outpipe\Client\OutpipeClient;
use Outpipe\Laravel\Console\HealthCommand;

final class OutpipeServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->mergeConfigFrom(__DIR__ . '/../../config/outpipe.php', 'outpipe');
        $this->app->singleton(OutpipeClient::class, static function (Application $app): OutpipeClient {
            $config = $app->make(Repository::class);

            return new OutpipeClient(
                (string) $config->get('outpipe.base_url'),
                $config->get('outpipe.api_key'),
                (float) $config->get('outpipe.timeout', 10),
            );
        });
        $this->app->alias(OutpipeClient::class, 'outpipe');
    }

    public function boot(): void
    {
        $this->publishes([
            __DIR__ . '/../../config/outpipe.php' => (string) $this->app->make('path.config') . '/outpipe.php',
        ], 'outpipe-config');

        if ($this->app->runningInConsole()) {
            $this->commands([HealthCommand::class]);
        }
    }
}
