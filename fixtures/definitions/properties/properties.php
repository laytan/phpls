<?php

namespace Test\Properties;

class TestPropertiesClass
{

    public $test = 'hello';

    function test()
    {
        return $this->test;
    }    

}

$testPropertiesObject = new TestPropertiesClass();
$testPropertiesVariable = $testPropertiesObject->test;
