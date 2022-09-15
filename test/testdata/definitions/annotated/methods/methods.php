<?php

namespace Definitions\Test\Methods;

class TestMethodsDecoyClass
{
    public function testPublic()
    {
    }
}

class TestMethodsClass extends TestTraitOverridesExtends  // @t_out(class_this, 1)
{
    public function testPublic() // @t_out(methods_public_same_class, 5) @t_out(methods_child_override_2, 5) @t_out(methods_on_newed_class_public, 5) @t_out(methods_at_var_phpdoc, 5) @t_out(methods_at_var_phpdoc_fqn, 5)
    {
    }

    protected function testProtected() // @t_out(methods_protected_same_class, 5) @t_out(methods_child_protected, 5)
    {
    }

    private function testPrivate() // @t_out(methods_private_same_class, 5)
    {
    }

    public function testMethodsSameClass()
    {
        $this->testPublic(); // @t_in(class_this, 10) @t_in(methods_public_same_class, 16)
        $this->testProtected(); // @t_in(methods_protected_same_class, 16)
        $this->testPrivate(); // @t_in(methods_private_same_class, 16)
    }

    public function testMethodToOverride()
    {
    }

    use TestMethodsTrait;

    public function testTraitFunctions()
    {
        $this->testTraitPublicFunction(); // @t_in(methods_trait_call, 16)
        $this->testTraitPrivateFunction(); // @t_in(methods_trait_private_call, 16)
        $this->testTraitOverridesExtends(); // @t_in(methods_trait_overrides_extends, 16)
    }
}

class TestTraitOverridesExtends
{
    public function testTraitOverridesExtends()
    {
    }
}


$testMethodsObject = new TestMethodsClass();
$testMethodsObject->testPublic(); // @t_in(methods_on_newed_class_public, 21)
$testMethodsObject->testProtected(); // @t_nodef(methods_on_newed_class_protected_nodef, 21)
$testMethodsObject->testPrivate();

$testMethodsObject->testTraitOverridesExtends(); // @t_in(methods_deep_trait, 21)

/**
 * @var TestMethodsClass
*/
$testphpdocmeth = $somekindofmagic;
$testphpdocmeth->testPublic(); // @t_in(methods_at_var_phpdoc, 18)

/**
 * @var \Definitions\Test\Methods\TestMethodsClass
*/
$testphpdocmethfqn = $somekindofmagic2;
$testphpdocmethfqn->testPublic(); // @t_in(methods_at_var_phpdoc_fqn, 21)
