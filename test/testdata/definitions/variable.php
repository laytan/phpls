<?php
// in: (7,8)
// out: (5,1)

$foo = 'bar';

echo $foo;

['foo' => $bar] = foobar();
echo $bar;

list('foo' => $foobar) = foobar();
echo $foobar;

[$foobarbaz] = foobar();
echo $foobarbaz;

list($foobarbazbaz) = foobar();
echo $foobarbazbaz;
