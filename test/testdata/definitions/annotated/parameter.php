<?php

$foo = '';
function foobar($foo) // @t_out(param_untyped_func, 17) @t_out(param_function_call, 1)
{
    return $foo; // @t_in(param_untyped_func, 13)
}

function closures($foo) // @t_out(param_decoy_closures, 19)
{
    $closure = function ($foo) { // @t_out(param_untyped_closure, 26)
        $foo; // @t_in(param_untyped_closure, 10)
    };

    $closure2 = function ($foo) { // @t_out(param_untyped_closure_2, 27)
        $foo; // @t_in(param_untyped_closure_2, 10)
    };

    $foo; // @t_in(param_decoy_closures, 6)

    $foo = fn($foo) => '.' . $foo; // @t_in(param_arrow_func, 30) @t_out(param_arrow_func, 15) @t_out(param_reassign, 5)

    echo $foo('test'); // @t_in(param_reassign, 11)
}

closures(foobar('')); // @t_in(param_function_call, 10)
