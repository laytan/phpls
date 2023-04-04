<?php
/**
 * A simple wrapper around PHP CS Fixer that runs it as a daemon.
 */

declare(strict_types=1);

require '../vendor/autoload.php';

use PhpCsFixer\Console\ConfigurationResolver;
use PhpCsFixer\Config;
use PhpCsFixer\ToolInfo;
use PhpCsFixer\Error\ErrorsManager;
use PhpCsFixer\Runner\Runner;
use PhpCsFixer\FileReader;

try {
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

    while (true) {
        $input = fgets(STDIN); // Read next line.
        if ($input === false) break; // Error happened, probably STDIN closed, so just exit.
        if (empty($input)) continue;

        file_put_contents('phpcsinputs', $input . PHP_EOL, FILE_APPEND);

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

        file_put_contents('phpcsoutputs', $changed['php://stdin']['diff'] . PHP_EOL, FILE_APPEND);

        echo json_encode($changed['php://stdin']['diff']) . PHP_EOL; // Just as string, but json encode so it's one line.
    }
} catch (\Throwable $e) {
    fwrite(STDERR, json_encode($e->getMessage()) . PHP_EOL);
}
