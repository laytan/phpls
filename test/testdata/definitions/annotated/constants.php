<?php // @t_dump(constants_dump, 0)

namespace Test\Constants;

define('CONSTANT_1', 'test'); // @t_out(constants_1, 1)

function test() {
    define('CONSTANT_1', 'test_overwrite'); // TODO: 2nd out for constants_1.
    define('CONSTANT_2', 'test2'); // @t_out(constants_2, 5)
    echo CONSTANT_2; // @t_in(constants_2, 10)
}

class TestConstants {

    public function test() {
        define('CONSTANT_3', 'test3'); // @t_out(constants_3, 9)
        echo CONSTANT_1; // @t_in(constants_1, 14)
    }
}

echo CONSTANT_3; // @t_in(constants_3, 6)
