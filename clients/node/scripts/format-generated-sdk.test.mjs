import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import {
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const scriptsDir = dirname(fileURLToPath(import.meta.url));
const packageRoot = resolve(scriptsDir, '..');
const formatterPath = join(scriptsDir, 'format-generated-sdk.mjs');
const prettierrcPath = join(packageRoot, '.prettierrc');

test('formats raw SDK trees that arrive through symlinked runfiles', () => {
  const tmp = mkdtempSync(join(tmpdir(), 'retab-node-sdk-format-'));

  try {
    const realRawDir = join(tmp, 'raw-real');
    const srcDir = join(realRawDir, 'src', 'nested');
    mkdirSync(srcDir, { recursive: true });
    writeFileSync(join(realRawDir, '.oagen-manifest.json'), '{"generatedAt":"test"}\n');

    const backingSource = join(tmp, 'backing-model.ts');
    writeFileSync(backingSource, "export const value={foo:'bar'}\n");
    symlinkSync(backingSource, join(srcDir, 'model.ts'));

    const rawRunfileLink = join(tmp, 'raw-runfile-link');
    symlinkSync(realRawDir, rawRunfileLink, 'dir');

    const outDir = join(tmp, 'formatted');
    const result = spawnSync(process.execPath, [formatterPath], {
      cwd: packageRoot,
      encoding: 'utf8',
      env: {
        ...process.env,
        RETAB_RAW_SDK_DIR: rawRunfileLink,
        RETAB_FORMATTED_SDK_DIR: outDir,
        RETAB_PRETTIERRC: prettierrcPath,
      },
    });

    assert.equal(result.status, 0, result.stderr || result.stdout);
    assert.match(result.stdout, /formatted 1\/1 TypeScript files/);

    const formattedPath = join(outDir, 'src', 'nested', 'model.ts');
    assert.equal(readFileSync(formattedPath, 'utf8'), "export const value = { foo: 'bar' };\n");
    assert.equal(lstatSync(formattedPath).isSymbolicLink(), false);
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }
});
