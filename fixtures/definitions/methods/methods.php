<?php

namespace Definitions\Test\Methods;

class TestMethodsDecoyClass
{
    public function testPublic()
    {
    }
}

class TestMethodsClass extends TestTraitOverridesExtends
{
    public function testPublic()
    {
    }

    protected function testProtected()
    {
    }

    private function testPrivate()
    {
    }

    public function testMethodsSameClass()
    {
        $this->testPublic();
        $this->testProtected();
        $this->testPrivate();
    }

    public function testMethodToOverride()
    {
    }

    use TestMethodsTrait;
    
    public function testTraitFunctions()
    {
        $this->testTraitPublicFunction();
        $this->testTraitPrivateFunction();
        $this->testTraitOverridesExtends();
    }
}

class TestTraitOverridesExtends
{
    public function testTraitOverridesExtends()
    {
    }
}
