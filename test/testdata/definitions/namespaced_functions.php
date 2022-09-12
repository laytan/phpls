<?php

function testing_namespaced_function_global_ok()
{
}

namespace TestingNamespacedFunctions1;

function test_namespaced_function_1()
{
}

namespace TestingNamespacedFunctions2;

testing_namespaced_function_global_ok();
test_namespaced_function_1();
