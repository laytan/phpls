<?php

namespace Test\ClassConstants;

interface TestConstInterface {
    public const INTERFACE = 'interface'; // @t_out(classconst_interface, 5) @t_out(classconst_interface_outside, 5)
}

class Test {
    public const FOO = 'test'; // @t_out(classconst_this, 5) @t_out(classconst_self, 5) @t_out(classconst_parent, 5) @t_out(classconst_static, 5) @t_out(classconst_direct, 5) @t_out(classconst_variable, 5)
    public const FOO_OUTSIDE = 'test'; // @t_out(classconst_direct_outside, 5) @t_out(classconst_direct_outside_2, 5) @t_out(classconst_variable_outside, 5)
    protected const PROTECTED = 'protected'; // @t_out(classconst_protected, 5)
    const NO_VISIBILITY = 'no-visibility'; // @t_out(classconst_no_visibility, 5) @t_out(classconst_no_visibility_outside, 5)
}

class Test2 extends Test {
    private const PRIVATE = 'private'; // @t_out(classconst_private, 5)

    public function test() {
        echo $this::FOO; // @t_in(classconst_this, 21)
        echo self::FOO; // @t_in(classconst_self, 20)
        echo parent::FOO; // @t_in(classconst_parent, 22)
        echo static::FOO; // @t_in(classconst_static, 22)
        echo Test2::FOO; // @t_in(classconst_direct, 21)

        $s = $this;
        echo $s::FOO; // @t_in(classconst_variable, 18)

        echo TestConstInterface::INTERFACE; // @t_in(classconst_interface, 34)
        echo Test::PROTECTED; // @t_in(classconst_protected, 20)
        echo Test::NO_VISIBILITY; // @t_in(classconst_no_visibility, 20)
        echo Test2::PRIVATE; // @t_in(classconst_private, 21)
    }
}

(new Test2())->test();

echo Test::FOO_OUTSIDE; // @t_in(classconst_direct_outside, 12)

echo Test2::FOO_OUTSIDE; // @t_in(classconst_direct_outside_2, 13)

$t = new Test2();
$t::FOO_OUTSIDE; // @t_in(classconst_variable_outside, 5)

echo TestConstInterface::INTERFACE; // @t_in(classconst_interface_outside, 26)
echo Test::PROTECTED; // @t_nodef(classconst_protected_outside, 12)
echo Test::NO_VISIBILITY; // @t_in(classconst_no_visibility_outside, 12)
echo Test2::PRIVATE; // @t_nodef(classconst_private_outside, 13)
