<?php
/**
 * A simple wrapper around PHP CS Fixer that runs it as a daemon.
 */

declare(strict_types=1);

// TODO: dynamic:
require '/Users/laytan/projects/elephp/tools/formatter/vendor/autoload.php';

use PhpCsFixer\Console\ConfigurationResolver;
use PhpCsFixer\Config;
use PhpCsFixer\ToolInfo;
use PhpCsFixer\Error\ErrorsManager;
use PhpCsFixer\Runner\Runner;
use PhpCsFixer\FileReader;

$resolver = new ConfigurationResolver(
    new Config(),
    [
        'path' => ['-'], // stdin
        'diff' => true,
        'format' => 'json',
        'stop-on-violation' => true,
    ],
    getcwd(),
    new ToolInfo(),
);

$errorsManager = new ErrorsManager();

$fileReader = FileReader::createSingleton();

$runner = new Runner(
    $resolver->getFinder(),
    $resolver->getFixers(),
    $resolver->getDiffer(),
    null,
    $errorsManager,
    $resolver->getLinter(),
    $resolver->isDryRun(),
    $resolver->getCacheManager(),
    $resolver->getDirectory(),
    $resolver->shouldStopOnViolation(),
);

$errs = invade($errorsManager);

// Make fgets below block.
stream_set_blocking(STDIN, true);

// TODO: try catch everything.
while (true) {
    $input = fgets(STDIN); // Read next line.
    if (empty($input)) {
        continue;
    }

    $input = json_decode($input);
    if (empty($input)) {
        fwrite(STDERR, json_encode('the input: "'.$input.'" is not valid JSON') . PHP_EOL);
        continue;
    }

    invade($fileReader)->stdinContent = $input; // Pretty cool hack to manipulate what PHP CS Fixer fixes.

    $changed = $runner->fix();

    if (count($errs->errors) > 0) {
        fwrite(STDERR, json_encode($errs->errors[0]->getSource()->getMessage()) . PHP_EOL); // Just as string, but json encode so it's one line.
        $errs->errors = [];
        continue;
    }

    echo json_encode($changed['php://stdin']['diff']) . PHP_EOL; // Just as string, but json encode so it's one line.
}
