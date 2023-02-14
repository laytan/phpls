<?php

function testing_namespaced_function_global_ok() // @t_out(func_namespaced, 1)
{
}

namespace TestingNamespacedFunctions1;

function test_namespaced_function_1()
{
}

namespace TestingNamespacedFunctions2;

testing_namespaced_function_global_ok(); // @t_in(func_namespaced, 1)
test_namespaced_function_1(); // @t_nodef(func_from_other_namespace_nodef, 1)
