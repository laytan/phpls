<?php

namespace Definitions\Test\Methods;

class MethodInheritDocTarget {

    public function target() {} // @t_out(methods_inheritdoc, 5)

    public string $id = "bleh"; // @t_out(methods_inheritdoc_to_prop, 5)
}

class MethodInheritDocParent {

    /**
     * @return MethodInheritDocTarget
     */
    public function test() {}
}

class MethodInheritDoc extends MethodInheritDocParent {

    /**
     * {@inheritdoc}
     */
    public function test() {}
}

$testInheritDoc = new MethodInheritDoc();
$testInheritDoc->test()->target(); // @t_in(methods_inheritdoc, 26)
$testInheritDoc->test()->id; // @t_in(methods_inheritdoc_to_prop, 26)
