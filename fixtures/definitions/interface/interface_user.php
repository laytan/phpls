<?php

use TestInterface\TestInterfaceInNamespaceInterface;

class TestInterfaceClassGlobal implements TestInterfaceInterface
{
}

class TestInterfaceClassNamespace implements TestInterfaceInNamespaceInterface
{
}

class TestInterfaceClassStdlib implements DateTimeInterface
{
}

interface TestInterfaceInlineInterface
{
}

class TestInterfaceClassInline implements TestInterfaceInlineInterface
{
}

class TestInterfaceClassMultiple implements TestInterfaceInlineInterface, DateTimeInterface
{
}
