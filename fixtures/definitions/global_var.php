<?php
$a = 1;

function foo()
{
    echo $a;
}

function bar()
{
    global $a;
    echo $a;
}
