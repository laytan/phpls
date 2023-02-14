<?php

use TestTrait\TestTraitTraitInNamespace;

class TestTraitUser
{
    use TestTraitTrait; // @t_in(trait_usage_other_file, 9)
    use TestTraitTraitInNamespace; // @t_in(trait_usage_other_file_and_namespace, 9)
}
