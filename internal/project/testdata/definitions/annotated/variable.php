<?php
$foo = 'bar'; // @t_out(var_basic_echo_string, 1)

echo $foo; // @t_in(var_basic_echo_string, 6)

['foo' => $bar] = foobar(); // @t_out(var_assoc_arr_unpack, 11)
echo $bar; // @t_in(var_assoc_arr_unpack, 6)

list('foo' => $foobar) = foobar(); // @t_out(var_assoc_list_unpack, 15)
echo $foobar; // @t_in(var_assoc_list_unpack, 6)

[$foobarbaz] = foobar(); // @t_out(var_arr_unpack, 2)
echo $foobarbaz; // @t_in(var_arr_unpack, 6)

list($foobarbazbaz) = foobar(); // @t_out(var_list_unpack, 6)
echo $foobarbazbaz; // @t_in(var_list_unpack, 6)
