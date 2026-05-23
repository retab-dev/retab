<?php

declare(strict_types=1);

use PhpCsFixer\Config;
use PhpCsFixer\Finder;

$finder = (new Finder())
    ->in(__DIR__)
    ->exclude(['vendor'])
    ->name('*.php');

return (new Config())
    ->setRiskyAllowed(false)
    ->setRules([
        '@PER-CS' => true,
        'declare_strict_types' => true,
    ])
    ->setFinder($finder);
