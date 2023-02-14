<?php

namespace Definitions\Test\Methods;

class ReturnStaticSelf {

    public string $target = 'bleh'; // @t_out(methods_return_self, 5) @t_out(methods_return_static, 5)

    /**
     * @return self
     */
    public function returnsSelf() {}

    /*
     * @return static
     */
    public function returnsStatic() {}
}

$testReturnStaticSelf = new ReturnStaticSelf();
$testReturnStaticSelf->returnsSelf()->target; // @t_in(methods_return_self, 39)
$testReturnStaticSelf->returnsStatic()->target; // @t_in(methods_return_static, 41)
