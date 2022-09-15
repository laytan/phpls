<?php

namespace Definitions\Test\Methods;

class TestMethodsChildClass extends TestMethodsClass
{
    public function testMethodsChildClass()
    {
        $this->testPublic(); // @t_in(methods_child_override_2, 16)
        $this->testProtected(); // @t_in(methods_child_protected, 16)
        $this->testPrivate(); // TODO: fix&enable: @t_skip_nodef(methods_child_private_nodef, 16)
        $this->testMethodToOverride(); // @t_in(methods_child_override, 16)
    }

    public function testMethodToOverride() // @t_out(methods_child_override, 5)
    {
    }

    public function testTraitUsageOfExtendedClass()
    {
        $this->testTraitOverridesExtends(); // @t_in(methods_child_trait_override, 16)
    }
}
