// Hermetic prettier-format step for the sandboxed Node SDK drift check.
//
// The generator (//factory/generators/oagen/retab-gen:generated_node_sdk_raw) emits an
// UNFORMATTED SDK tree (src/ + .oagen-manifest.json) in a network-free Bazel
// sandbox. Prettier is NOT part of the generator's dependency closure, so this
// downstream step formats that raw tree with the Node SDK client's OWN vendored
// prettier (node_sdk_npm, pinned to the version in this package's
// package-lock.json) plus this package's .prettierrc — matching exactly how the
// committed SDK was formatted. The output is a formatted tree that the drift
// check then diffs against the committed src/ + manifest.
//
// Paths are passed via env:
//   RETAB_RAW_SDK_DIR    - $(execpath) of the raw generated tree (execroot-rel).
//   RETAB_FORMATTED_SDK_DIR - the declared out_dir. Plain relative values are
//                          resolved against cwd because js_run_binary chdirs to
//                          the Bazel package output directory before execution.
//   RETAB_PRETTIERRC     - $(execpath) of this package's .prettierrc.
// Relative execpath values are resolved against JS_BINARY__EXECROOT, which
// aspect_rules_js sets to the absolute execroot even when chdir is in effect.
import {
  chmodSync,
  cpSync,
  lstatSync,
  readFileSync,
  readdirSync,
  rmSync,
  statSync,
  writeFileSync,
} from 'node:fs';
import { isAbsolute, join, resolve } from 'node:path';
import prettier from 'prettier';

function fromExecroot(value) {
  if (!value) {
    throw new Error('missing required path env var');
  }
  if (isAbsolute(value)) {
    return value;
  }
  const execroot = process.env.JS_BINARY__EXECROOT;
  return resolve(execroot ?? process.cwd(), value);
}

function outputDirPath(value) {
  const output = value ?? 'generated-node-formatted';
  if (isAbsolute(output)) {
    return output;
  }
  const execroot = process.env.JS_BINARY__EXECROOT;
  const bindir = process.env.BAZEL_BINDIR;
  if (execroot && bindir && (output === bindir || output.startsWith(`${bindir}/`))) {
    return resolve(execroot, output);
  }
  return resolve(output);
}

const rawDir = fromExecroot(process.env.RETAB_RAW_SDK_DIR);
const outDir = outputDirPath(process.env.RETAB_FORMATTED_SDK_DIR);
const prettierrcPath = fromExecroot(process.env.RETAB_PRETTIERRC);

function makeWritableTree(path) {
  const metadata = lstatSync(path);
  if (metadata.isSymbolicLink()) {
    return;
  }
  if (metadata.isDirectory()) {
    chmodSync(path, 0o755);
    for (const entry of readdirSync(path)) {
      makeWritableTree(join(path, entry));
    }
    return;
  }
  chmodSync(path, 0o644);
}

function removeOutputTree(path) {
  try {
    rmSync(path, { recursive: true, force: true });
  } catch (error) {
    if (error && ['EACCES', 'ENOTEMPTY', 'EPERM'].includes(error.code)) {
      makeWritableTree(path);
      rmSync(path, { recursive: true, force: true });
      return;
    }
    throw error;
  }
}

// Mirror the raw tree (src/ + .oagen-manifest.json) into the declared output.
removeOutputTree(outDir);
cpSync(rawDir, outDir, { dereference: true, recursive: true });

const prettierConfig = JSON.parse(readFileSync(prettierrcPath, 'utf8'));

function collectTsFiles(dir) {
  const out = [];
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) {
      out.push(...collectTsFiles(full));
    } else if ((entry.isFile() || entry.isSymbolicLink()) && entry.name.endsWith('.ts')) {
      if (!statSync(full).isFile()) {
        continue;
      }
      out.push(full);
    }
  }
  return out;
}

const srcDir = join(outDir, 'src');
const files = collectTsFiles(srcDir).sort();

let formattedCount = 0;
for (const file of files) {
  const source = readFileSync(file, 'utf8');
  // `filepath` makes prettier infer the TypeScript parser exactly like the CLI
  // (`prettier --write "src/**/*.ts"`); the .prettierrc options are applied on
  // top so the output is byte-identical to the committed, hand-run formatting.
  const formatted = await prettier.format(source, { ...prettierConfig, filepath: file });
  const fileMetadata = lstatSync(file);
  if (formatted !== source || fileMetadata.isSymbolicLink()) {
    // cpSync preserves the read-only mode of the Bazel-declared raw tree inputs,
    // so make the copy writable before overwriting it. If the copied path is a
    // runfiles symlink, replace the link itself instead of mutating its target.
    if (fileMetadata.isSymbolicLink()) {
      rmSync(file);
    } else {
      rmSync(file, { force: true });
    }
    writeFileSync(file, formatted);
  }
  if (formatted !== source) {
    formattedCount += 1;
  }
}

console.log(`[format-generated-sdk] formatted ${formattedCount}/${files.length} TypeScript files`);
