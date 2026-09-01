<?php

namespace Outpipe\Client;

use Outpipe\Resources\Agent;
use Outpipe\Resources\ApiKey;
use Outpipe\Resources\Collection;
use Outpipe\Resources\Domain;
use Outpipe\Resources\Organization;
use Outpipe\Resources\Resource;
use Outpipe\Resources\Tunnel;
use Outpipe\Resources\Webhook;

final class ResourceMapper
{
    public static function one(array $data, string $type): Resource
    {
        return new $type($data);
    }

    public static function many(array $data, string $type): Collection
    {
        $items = array_map(static fn (array $item): Resource => new $type($item), $data);

        return new Collection($items);
    }

    public static function organization(array $data): Organization { return new Organization($data); }
    public static function tunnel(array $data): Tunnel { return new Tunnel($data); }
    public static function agent(array $data): Agent { return new Agent($data); }
    public static function domain(array $data): Domain { return new Domain($data); }
    public static function apiKey(array $data): ApiKey { return new ApiKey($data); }
    public static function webhook(array $data): Webhook { return new Webhook($data); }
}
