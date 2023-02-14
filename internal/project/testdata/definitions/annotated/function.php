<?php

function foo() // @t_out(func_untyped_unparamed, 1)
{
}

foo(); // @t_in(func_untyped_unparamed, 1)

function barfunction()
{
    function foobarfunction() // @t_out(func_inside_func, 5)
    {
    }

    foobarfunction(); // @t_in(func_inside_func, 5)
}

foobarfunction(); // @t_nodef(func_undefined, 1)

