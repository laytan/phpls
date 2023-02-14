<?php
// This causes a panic while parsing, it should continue.
while (false)
    echo 'faiL';

do
    echo 'dowhile test';
while (false);

    echo 'dowhile test';
} while (false);

while (false) {
    echo 'pass';
}

do {
