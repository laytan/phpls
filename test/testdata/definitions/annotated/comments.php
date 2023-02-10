<?php

namespace Test\Annotated\Comments;

class Exception {} // @t_out(comments_throws, 1)

class TestClass {} // @t_out(comments_returns, 1)

/**
 * @throws Exception // @t_in(comments_throws, 13)
 * @return \Test\Annotated\Comments\TestClass :) // @t_in(comments_returns, 39)
 */
function test() {}

