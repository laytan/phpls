<?php

namespace Definitions\Test\Methods;

class TestMethodsChildClass extends TestMethodsClass
{
    public function testMethodsChildClass()
    {
        $this->testPublic();
        $this->testProtected();
        $this->testPrivate(); 
        $this->testMethodToOverride();
    }

    public function testMethodToOverride()
    {
    }

    public function testTraitUsageOfExtendedClass()
    {
        $this->testTraitOverridesExtends();
    }
}
