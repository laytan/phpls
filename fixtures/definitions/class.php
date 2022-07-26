<?php

use Swoole\WebSocket\Server;
use Swoole\Process as SwooleProcess;

namespace TestClass;

class DateTimeImmutable
{
}

$timestamp = (new \DateTimeImmutable())
    ->getTimestamp();

$localInstance = new DateTimeImmutable();

$server = new Server();

$process = new SwooleProcess();

