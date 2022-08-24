<?php
// in: (7,12)
// out: (5,17)
$foo = '';
function foobar($foo)
{
    return $foo;
}

function closures($foo)
{
    $closure = function ($foo) {
        $foo;
    };

    $closure2 = function ($foo) {
        $foo;
    };

    $foo;
}
