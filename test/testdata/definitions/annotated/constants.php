<?php

namespace Test\Constants;

define('CONSTANT_1', 'test'); // @t_out(constants_1, 1)
define('CONSTANT_2', 'test2'); // @t_out(constants_2, 1)

function test() {
    echo CONSTANT_2; // @t_in(constants_2, 10)
}

define('CONSTANT_1', 'test_overwrite'); // @t_out(constants_1, 1)

class TestConstants {

    public function test() {
        echo CONSTANT_1; // @t_in(constants_1, 14)
    }
}

define('CONSTANT_3', 'test3'); // @t_out(constants_3, 1)

echo CONSTANT_3; // @t_in(constants_3, 6)

/**
 * This is currently not supported, this is rarely done and supporting this
 * would mean we have to walk every node, in the symbol traverser (indexer).
 * Resulting in many more calls for this rarely used feature.
 */
function testInScope() {
    define('CONSTANT_IN_SCOPE', 'test_in_scope');
}
echo CONSTANT_IN_SCOPE; // @t_nodef(constants_in_scope, 6)
