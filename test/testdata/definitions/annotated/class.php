<?php

namespace TestClass;

class DateTimeImmutable // @t_out(class_local_same_name_as_global, 1)
{
}

$localInstance = new DateTimeImmutable(); // @t_in(class_local_same_name_as_global, 22)
