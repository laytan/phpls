<?php

namespace Test\Annotated\Comments;

class Exception {} // @t_out(comments_throws, 1) @t_out(comments_param_method, 1) @t_out(comments_param_method_public_static, 1) @t_out(comments_return_method_final, 1) @t_out(comments_final_class, 1)

/**
 * @var Exception // @t_in(comments_final_class, 11)
 */
final class TestClass {} // @t_out(comments_returns, 1) @t_out(comments_property_var, 1) @t_out(comments_class, 1)

/**
 * @throws Exception // @t_in(comments_throws, 13)
 * @return \Test\Annotated\Comments\TestClass :) // @t_in(comments_returns, 39)
 */
function test() {}

/**
 * @var TestClass // @t_in(comments_class, 12)
 */
class MethodTest { // @t_out(comments_trait, 1) @t_out(comments_interface, 1)

    /**
     * @var TestClass // @t_in(comments_property_var, 16)
     */
    public $test = 'test';

    /**
     * @param Exception $a // @t_in(comments_param_method, 19)
     */
    function test1($a) {} // @t_out(comments_inherit, 5)

    /**
     * @param Exception $a // @t_in(comments_param_method_public_static, 19)
     */
    public static function test2($a) {}

    /**
     * @return Exception // @t_in(comments_return_method_final, 19)
     */
    final function test3() {}
}

/**
 * @var MethodTest // @t_in(comments_trait, 11)
 */
trait TraitTest {}

/**
 * @var MethodTest // @t_in(comments_interface, 11)
 */
interface InterfaceTest {}

class InheritsTest extends MethodTest {

    /**
     * {@inheritdoc} // @t_in(comments_inherit, 12)
     */
    public function test1($a) {}
}
