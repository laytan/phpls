<?php

use TestInterface\TestInterfaceInNamespaceInterface;

class TestInterfaceClassGlobal implements TestInterfaceInterface // @t_in(interface_global_implements, 43)
{
}

class TestInterfaceClassNamespace implements TestInterfaceInNamespaceInterface // @t_in(interface_namespaced_implements, 46)
{
}

interface TestInterfaceInlineInterface // @t_out(interface_same_file_global, 1) @t_out(interface_implement_multiple, 1)
{
}

class TestInterfaceClassInline implements TestInterfaceInlineInterface // @t_in(interface_same_file_global, 43)
{
}

class TestInterfaceClassMultiple implements TestInterfaceInlineInterface, DateTimeInterface // @t_in(interface_implement_multiple, 45)
{
}
