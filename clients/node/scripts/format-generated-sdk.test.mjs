import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import {
  chmodSync,
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

test('replaces existing read-only output trees', () => {
  const tmp = mkdtempSync(join(tmpdir(), 'retab-node-sdk-readonly-output-'));

  try {
    const rawDir = join(tmp, 'raw');
    const srcDir = join(rawDir, 'src');
    mkdirSync(srcDir, { recursive: true });
    writeFileSync(join(rawDir, '.oagen-manifest.json'), '{"generatedAt":"test"}\n');
    writeFileSync(join(srcDir, 'model.ts'), "export const value={foo:'bar'}\n");

    const outDir = join(tmp, 'formatted');
    const staleDir = join(outDir, 'src');
    mkdirSync(staleDir, { recursive: true });
    writeFileSync(join(staleDir, 'stale.ts'), 'export const stale = true;\n');
    chmodSync(join(staleDir, 'stale.ts'), 0o444);
    chmodSync(staleDir, 0o555);
    chmodSync(outDir, 0o555);

    const result = spawnSync(process.execPath, [formatterPath], {
      cwd: packageRoot,
      encoding: 'utf8',
      env: {
        ...process.env,
        RETAB_RAW_SDK_DIR: rawDir,
        RETAB_FORMATTED_SDK_DIR: outDir,
        RETAB_PRETTIERRC: prettierrcPath,
      },
    });

    assert.equal(result.status, 0, result.stderr || result.stdout);
    assert.equal(
      readFileSync(join(outDir, 'src', 'model.ts'), 'utf8'),
      "export const value = { foo: 'bar' };\n"
    );
  } finally {
    chmodSync(tmp, 0o755);
    rmSync(tmp, { recursive: true, force: true });
  }
});

test('writes plain relative out dirs under the current package output directory', () => {
  const tmp = mkdtempSync(join(tmpdir(), 'retab-node-sdk-package-out-'));

  try {
    const rawDir = join(tmp, 'raw');
    const srcDir = join(rawDir, 'src');
    mkdirSync(srcDir, { recursive: true });
    writeFileSync(join(rawDir, '.oagen-manifest.json'), '{"generatedAt":"test"}\n');
    writeFileSync(join(srcDir, 'model.ts'), "export const value={foo:'bar'}\n");

    const execroot = join(tmp, 'execroot');
    const bindir = 'bazel-out/darwin_arm64-fastbuild/bin';
    const packagePath = 'public/sdk/clients/node';
    const packageOutDir = join(execroot, bindir, packagePath);
    mkdirSync(packageOutDir, { recursive: true });

    const result = spawnSync(process.execPath, [formatterPath], {
      cwd: packageOutDir,
      encoding: 'utf8',
      env: {
        ...process.env,
        BAZEL_BINDIR: bindir,
        BAZEL_PACKAGE: packagePath,
        JS_BINARY__EXECROOT: execroot,
        RETAB_RAW_SDK_DIR: rawDir,
        RETAB_FORMATTED_SDK_DIR: 'generated-node-formatted',
        RETAB_PRETTIERRC: prettierrcPath,
      },
    });

    assert.equal(result.status, 0, result.stderr || result.stdout);

    const formattedPath = join(packageOutDir, 'generated-node-formatted', 'src', 'model.ts');
    assert.equal(readFileSync(formattedPath, 'utf8'), "export const value = { foo: 'bar' };\n");
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }
});

test('resolves execroot-relative Bazel output dirs from the execroot', () => {
  const tmp = mkdtempSync(join(tmpdir(), 'retab-node-sdk-execroot-out-'));

  try {
    const rawDir = join(tmp, 'raw');
    const srcDir = join(rawDir, 'src');
    mkdirSync(srcDir, { recursive: true });
    writeFileSync(join(rawDir, '.oagen-manifest.json'), '{"generatedAt":"test"}\n');
    writeFileSync(join(srcDir, 'model.ts'), "export const value={foo:'bar'}\n");

    const execroot = join(tmp, 'execroot');
    const bindir = 'bazel-out/darwin_arm64-fastbuild/bin';
    const packagePath = 'public/sdk/clients/node';
    const packageOutDir = join(execroot, bindir, packagePath);
    mkdirSync(packageOutDir, { recursive: true });

    const result = spawnSync(process.execPath, [formatterPath], {
      cwd: packageOutDir,
      encoding: 'utf8',
      env: {
        ...process.env,
        BAZEL_BINDIR: bindir,
        BAZEL_PACKAGE: packagePath,
        JS_BINARY__EXECROOT: execroot,
        RETAB_RAW_SDK_DIR: rawDir,
        RETAB_FORMATTED_SDK_DIR: `${bindir}/${packagePath}/generated-node-formatted`,
        RETAB_PRETTIERRC: prettierrcPath,
      },
    });

    assert.equal(result.status, 0, result.stderr || result.stdout);

    const formattedPath = join(packageOutDir, 'generated-node-formatted', 'src', 'model.ts');
    assert.equal(readFileSync(formattedPath, 'utf8'), "export const value = { foo: 'bar' };\n");
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }
});
