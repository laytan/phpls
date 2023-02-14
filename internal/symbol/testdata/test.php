<?php

new Test();

interface TestInterface {}

interface TestInterface2 {}

interface TestBaseInterface {}

class TestBase implements TestBaseInterface {}

trait TestTrait {
    function test() {}
}

class Test extends TestBase implements TestInterface, TestInterface2 {
    use TestTrait;
}
