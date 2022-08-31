<?php

// This namespace is defined in a file with multiple namespaces.
// I assumed this was invalid but it's actually not.
use http\Env;

class TestMultipleNamespacesInOneFile extends Env
{
}
