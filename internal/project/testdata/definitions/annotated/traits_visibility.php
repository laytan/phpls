<?php

trait TestTraitVisibilityTrait {
    private function test() {} // @t_out(traitvisibility_same_use_private_ok, 5)
    private string $testProp = 'tst'; // @t_out(traitvisibility_same_use_private_prop_ok, 5)
}

class TestTraitVisibility {
    use TestTraitVisibilityTrait;

    public function testYes() {
        $this->test(); // @t_in(traitvisibility_same_use_private_ok, 16)
        $this->testProp; // @t_in(traitvisibility_same_use_private_prop_ok, 16)
    }
}

class TestTraitVisibilityChild extends TestTraitVisibility {
    public function testNo() {
        $this->test(); // @t_nodef(traitvisibility_parent_use_private_no, 16)
        $this->testProp; // @t_nodef(traitvisibility_parent_use_private_prop_no, 16)
    }
}

