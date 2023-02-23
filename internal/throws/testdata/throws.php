<?php

namespace Throws\TestData;

function test_throws() {
    throw new \Exception("Hello World!");
}

function test_throws_2() {
    test_throws();

    throw new \Throwable("Hello World!");
}

function test_throws_3() {
    try {
        throw new \Exception("Hello World!");
    } catch (\Exception $e) {
    }
}

function test_throws_4() {
    try {
        throw new \Exception("Hello World!");
    } catch (\Throwable $e) {
    }
}

function test_throws_5() {
    try {
    } catch (\Exception $e) {
    }

    throw new \Exception("Hello World!");
}

/**
 * Example stub, which uses a tag and not any code.
 *
 * @throws \Throwable
 */
function test_throws_6() {}

function test_throws_7() {
    test_throws_6();
}

function test_throws_8() {
    throw new \InvalidArgumentException();
    throw new \InvalidArgumentException();
}

function test_throws_9() {
    try {
        throw new \InvalidArgumentException();
        throw new \Exception();
    } catch (\Exception $e) {
    }
}

function test_throws_10() {
    try {
    } catch (\Exception $e) {
        throw new \Exception("Hello World!");
    }
}

function test_throws_11() {
    try {
    } catch(\InvalidArgumentException $e) {
    } catch(\Throwable $e) {
        throw new \InvalidArgumentException();
    } finally {
        throw new \Exception();
    }
}
