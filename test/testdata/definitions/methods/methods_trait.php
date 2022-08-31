<?php

namespace Definitions\Test\Methods;

trait TestMethodsTrait
{
    public function testTraitPublicFunction()
    {
    }

    private function testTraitPrivateFunction()
    {
        $this->testTraitPublicFunction();
    }

    public function testTraitOverridesExtends()
    {
    }
}
