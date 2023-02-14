<?php

namespace Definitions\Test\Methods;

trait TestMethodsTrait
{
    public function testTraitPublicFunction() // @t_out(methods_trait_call_in_trait, 5) @t_out(methods_trait_call, 5)
    {
    }

    private function testTraitPrivateFunction() // @t_out(methods_trait_private_call, 5)
    {
        $this->testTraitPublicFunction(); // @t_in(methods_trait_call_in_trait, 16)
    }

    public function testTraitOverridesExtends() // @t_out(methods_trait_overrides_extends, 5) @t_out(methods_child_trait_override, 5) @t_out(methods_deep_trait, 5)
    {
    }
}
