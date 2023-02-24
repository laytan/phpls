<?php

namespace Throws\TestData\Classes;

class Thrown implements \Throwable {} // @t_out(throws_simple_parent, 1) @t_out(throws_simple_this, 1) @t_out(throws_simple_method, 1)

class Test {
    public function testing() { // @t_in(throws_simple_method, 5)
        throw new Thrown();
    }

    public function test_2() { // @t_in(throws_simple_this, 5)
        $this->testing();
    }
}

class TestExtended extends Test {
    public function testing() { // @t_in(throws_simple_parent, 5)
        parent::testing();
    }
}

