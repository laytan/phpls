<?php

namespace Test\Alias\Source;

class Url {} // @t_out(alias_double_classes, 1) @t_out(alias_normal_use, 1)

namespace Test\Alias\Source\Two;

class Url {} // @t_out(alias_aliassed, 1) @t_out(alias_use_definition, 1)

namespace Test\Alias;

use Test\Alias\Source\Url; // @t_in(alias_normal_use, 23)
use Test\Alias\Source\Two\Url as TwoUrl; // @t_in(alias_use_definition, 27)

new Url(); // @t_in(alias_double_classes, 5)
new TwoUrl(); // @t_in(alias_aliassed, 5)
