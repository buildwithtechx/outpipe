<?php

return [
    'base_url' => getenv('OUTPIPE_API_URL') ?: 'https://api.outpipe.dev',
    'api_key' => getenv('OUTPIPE_API_KEY') ?: null,
    'webhook_secret' => getenv('OUTPIPE_WEBHOOK_SECRET') ?: null,
    'timeout' => (float) (getenv('OUTPIPE_TIMEOUT') ?: 10),
];
