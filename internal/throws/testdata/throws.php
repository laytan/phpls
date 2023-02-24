<?php

namespace Throws\TestData;

interface Throwable extends \Throwable {} // @t_out(throws_recursive_func, 1) @t_out(throws_recursive_phpdoc, 1)
class Exception extends Throwable {} // @t_out(throws_simple_func, 1) @t_out(throws_recursive_func, 1) @t_out(throws_outside_of_catch, 1) @t_out(throws_rethrow, 1) @t_out(throws_rethrow_and_finally, 1)
class LogicException extends Exception {}
class InvalidArgumentException extends LogicException {} // @t_out(throws_2_same, 1) @t_out(throws_rethrow_and_finally, 1)

function test_throws() { // @t_in(throws_simple_func, 1)
    throw new Exception("Hello World!");
}

function test_throws_2() { // @t_in(throws_recursive_func, 1)
    test_throws();

    throw new Throwable("Hello World!");
}

function test_throws_3() { // @t_nodef(throws_catched, 1)
    try {
        throw new \Throws\TestData\Exception("Hello World!");
    } catch (Exception $e) {
    }
}

function test_throws_4() { // @t_nodef(throws_catched_2, 1)
    try {
        throw new Exception("Hello World!");
    } catch (Throwable $e) {
    }
}

function test_throws_5() { // @t_in(throws_outside_of_catch, 1)
    try {
    } catch (Exception $e) {
    }

    throw new Exception("Hello World!");
}

/**
 * Example stub, which uses a tag and not any code.
 *
 * @throws \Throws\TestData\Throwable
 */
function test_throws_6() {}

function test_throws_7() { // @t_in(throws_recursive_phpdoc, 1)
    test_throws_6();
}

function test_throws_8() { // @t_in(throws_2_same, 1)
    throw new InvalidArgumentException();
    throw new InvalidArgumentException();
}

function test_throws_9() { // @t_nodef(throws_caught_overarching, 1)
    try {
        throw new InvalidArgumentException();
        throw new Exception();
    } catch (Exception $e) {
    }
}

function test_throws_10() { // @t_in(throws_rethrow, 1)
    try {
    } catch (Exception $e) {
        throw new Exception("Hello World!");
    }
}

function test_throws_11() { // @t_in(throws_rethrow_and_finally, 1)
    try {
    } catch(InvalidArgumentException $e) {
    } catch(Throwable $e) {
        throw new InvalidArgumentException();
    } finally {
        throw new Exception();
    }
}

