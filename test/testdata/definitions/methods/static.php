<?php

class TestStaticMethods
{

    public static function testPublic()
    {
    }

    protected static function testProtected()
    {
    }

    private static function testPrivate()
    {
    }
}

class TestStaticMethodsChild extends TestStaticMethods
{
}

TestStaticMethods::testPublic();
TestStaticMethods::testProtected();
TestStaticMethods::testPrivate();

TestStaticMethodsChild::testPublic();
TestStaticMethodsChild::testProtected();
TestStaticMethodsChild::testPrivate();

$test = new TestStaticMethods();
$test::testPublic();
