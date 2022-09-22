<?php

class ExprTest2 {} // @t_out(complex_func_ret_hint_property_hint, 1) @t_out(basic_at_var_doc_property, 1)
class ExprTest3 {} // @t_out(basic_static_doc, 1) @t_out(complex_var_static_doc, 1)

class ExprTest { // @t_out(basic_func_ret_hint, 1)

    public ExprTest2 $testPropertyHint;

    /** @return ExprTest3 */
    public static function testStaticDoc() {}
}

function exprTest(): ExprTest {}

exprTest(); // @t_in(basic_func_ret_hint, 1)
exprTest()->testPropertyHint; // @t_in(complex_func_ret_hint_property_hint, 13)

/** @var ExprTest */
$test = null;
$test->testPropertyHint; // @t_in(basic_at_var_doc_property, 8)

ExprTest::testStaticDoc(); // @t_in(basic_static_doc, 11)

$test::testStaticDoc(); // @t_in(complex_var_static_doc, 8)

