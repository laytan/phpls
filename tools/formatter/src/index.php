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
	$cwd = getcwd();

	// TODO: automatically set indentation configuration based on passed
	// in params in the LSP format request.
	$customConfig = (new ConfigurationResolver(new Config(), ['path' => ['-']], $cwd, new ToolInfo()))
		->getConfig(); // This will return the config in cwd if there is one.

    $resolver = new ConfigurationResolver(
        $customConfig, // Use the user's config as a default.
        [ // Overwrite the necessary config for the daemon.
            'path' => ['-'], // stdin
            'diff' => true,
            'format' => 'json',
            'stop-on-violation' => true,
        ],
        '/ajskdhaskbscajskdjashascbcnajdk', // just a non existing path.
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

		if (!isset($changed['php://stdin']['diff'])) {
			$changed['php://stdin']['diff'] = '';
		}

        echo json_encode($changed['php://stdin']['diff']) . PHP_EOL; // Just as string, but json encode so it's one line.
    }
} catch (\Throwable $e) {
	$msg = sprintf('%s:%s - %s', $e->getFile(), $e->getLine(), $e->getMessage());
    fwrite(STDERR, json_encode($msg . PHP_EOL));
}
