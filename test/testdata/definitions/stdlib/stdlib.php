<?php

use Swoole\WebSocket\Server;
use Swoole\Process as SwooleProcess;

// This namespace is defined in a file with multiple namespaces.
// I assumed this was invalid but it's actually not.
use http\Env;

echo in_array('foo', []);

$timestamp = (new \DateTimeImmutable())
    ->getTimestamp();

$server = new Server();

$process = new SwooleProcess();

class TestDateTimeImplements implements DateTimeInterface {}

class TestMultipleNamespacesInOneFile extends Env
{
}

closures(trim(''));
