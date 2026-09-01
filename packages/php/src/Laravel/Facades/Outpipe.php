<?php

namespace Outpipe\Laravel\Facades;

use Illuminate\Support\Facades\Facade;

final class Outpipe extends Facade
{
    protected static function getFacadeAccessor(): string
    {
        return 'outpipe';
    }
}
