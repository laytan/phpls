<?php

namespace Test\Properties;

class TestPropertiesClass
{
    public $test = 'hello';
    protected $test3 = 'hello';
    private $test2 = 'hello';

    function test()
    {
        return $this->test;
    }    
}

$testPropertiesObject = new TestPropertiesClass();
$testPropertiesVariable = $testPropertiesObject->test;
$testPropertiesObject->test2;
$testPropertiesObject->test3;

/**
 * @var TestPropertiesClass 
*/
$testphpdocprop = $somekindofmagic;
$testphpdocprop->test;

/**
 * @var \Test\Properties\TestPropertiesClass 
*/
$testphpdocpropfqn = $somekindofmagic2;
$testphpdocpropfqn->test;


class TestPropertiesClassDummy
{
}

class TestPropertiesClass2
{
    public TestPropertiesClass $testType;

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
$testingProperties->testType->test;
$testingProperties->testDoc->test;
$testingProperties->testDocAndType->test;
$testingProperties->testType->nonexistantprop;
$testingProperties->testType->test3;
$testingProperties->testPrivate->test;
