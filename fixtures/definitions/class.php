<?php

use Swoole\WebSocket\Server;

namespace TestClass;

class DateTimeImmutable
{
}

$timestamp = (new \DateTimeImmutable())
    ->getTimestamp();

$localInstance = new DateTimeImmutable();

$server = new Server();

