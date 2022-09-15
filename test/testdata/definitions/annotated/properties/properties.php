<?php

namespace Test\Properties;

class TestPropertiesClass // @t_out(class_this_properties, 1)
{
    public $test = 'hello'; // @t_out(props_same_class_public, 5) @t_out(props_on_newed_class_public, 5) @t_out(props_on_at_var_public, 5) @t_out(props_on_at_var_fqn_public, 5) @t_out(props_chained_typehint, 5) @t_out(props_chained_phpdoc, 5) @t_out(props_chained_doc_and_hint, 5) @t_out(props_chained_child, 5)
    protected $test3 = 'hello';
    private $test2 = 'hello';

    function test()
    {
        return $this->test; // @t_in(class_this_properties, 17) @t_in(props_same_class_public, 23)
    }
}

$testPropertiesObject = new TestPropertiesClass();
$testPropertiesVariable = $testPropertiesObject->test; // @t_in(props_on_newed_class_public, 50)
$testPropertiesObject->test2; // @t_nodef(props_on_newed_class_protected_nodef, 24)
$testPropertiesObject->test3; // @t_nodef(props_on_newed_class_private_nodef, 24)

/**
 * @var TestPropertiesClass
*/
$testphpdocprop = $somekindofmagic;
$testphpdocprop->test; // @t_in(props_on_at_var_public, 18)

/**
 * @var \Test\Properties\TestPropertiesClass
*/
$testphpdocpropfqn = $somekindofmagic2;
$testphpdocpropfqn->test; // @t_in(props_on_at_var_fqn_public, 21)


class TestPropertiesClassDummy
{
}

class TestPropertiesClass2
{
    public TestPropertiesClass $testType; // @t_out(props_chained_def_middle, 5)

    /**
     * @var \Test\Properties\TestPropertiesClass
     */
    public $testDoc;

    /**
     * @var \Test\Properties\TestPropertiesClass
     */
    public TestPropertiesClassDummy $testDocAndType;

    private TestPropertiesClass $testPrivate;
}

$testingProperties = new TestPropertiesClass2();
$testingProperties->testType->test; // @t_in(props_chained_typehint, 31)
$testingProperties->testDoc->test; // @t_in(props_chained_phpdoc, 30)
$testingProperties->testDocAndType->test; // @t_in(props_chained_doc_and_hint, 37)
$testingProperties->testType->nonexistantprop; // @t_nodef(props_chained_nodef, 31)
$testingProperties->testType->test3; // @t_nodef(props_chained_private_end_nodef, 31)
$testingProperties->testPrivate->test; // @t_nodef(props_chained_private_start_nodef, 34)

class TestPropertiesClassChild extends TestPropertiesClass2
{
}

$testingPropertiesChild = new TestPropertiesClassChild();
$testingPropertiesChild->testType->test; // @t_in(props_chained_def_middle, 26) @t_in(props_chained_child, 36)
$testingPropertiesChild->testNo->test; // @t_nodef(props_chained_def_middle_nodef, 36) // @t_nodef(props_chained_def_nodef, 34)

class TestPropertiesClassReturnedByMethod
{
    public TestPropertiesClass $testType; // @t_out(props_chained_with_at_varred_methods, 5) @t_out(props_chained_with_typehinted_methods, 5)
}

class TestPropertiesClassChild2 extends TestPropertiesClass2
{

    /**
     * @return TestPropertiesClassReturnedByMethod
     */
    public function test()
    {
    }

    public function test2(): TestPropertiesClassReturnedByMethod
    {
    }
}

$testChild2 = new TestPropertiesClassChild2();
$testChild2->test()->testType; // @t_in(props_chained_with_at_varred_methods, 21)
$testChild2->test2()->testType; // @t_in(props_chained_with_typehinted_methods, 22)

class TestTypedProperties
{
    public string $testTypedProperty; // @t_out(props_hint_non_classlike, 5)
}

$testTypedPropObj = new TestTypedProperties();
$testTypedPropObj->testTypedProperty; // @t_in(props_hint_non_classlike, 20)
