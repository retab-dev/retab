// Hermetic prettier-format step for the sandboxed Node SDK drift check.
//
// The generator (//.oagen-workspace/retab-gen:generated_node_sdk_raw) emits an
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
//   RETAB_FORMATTED_SDK_DIR - the declared out_dir (resolved against cwd; the
//                          js_run_binary runs with chdir = this package).
//   RETAB_PRETTIERRC     - $(execpath) of this package's .prettierrc.
// Relative execpath values are resolved against JS_BINARY__EXECROOT, which
// aspect_rules_js sets to the absolute execroot even when chdir is in effect.
import {
  chmodSync,
  cpSync,
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

const rawDir = fromExecroot(process.env.RETAB_RAW_SDK_DIR);
const outDir = resolve(process.env.RETAB_FORMATTED_SDK_DIR ?? 'generated-node-formatted');
const prettierrcPath = fromExecroot(process.env.RETAB_PRETTIERRC);

// Mirror the raw tree (src/ + .oagen-manifest.json) into the declared output.
rmSync(outDir, { recursive: true, force: true });
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
  if (formatted !== source) {
    // cpSync preserves the read-only mode of the Bazel-declared raw tree inputs,
    // so make the copy writable before overwriting it with the formatted output.
    chmodSync(file, 0o644);
    writeFileSync(file, formatted);
    formattedCount += 1;
  }
}

console.log(`[format-generated-sdk] formatted ${formattedCount}/${files.length} TypeScript files`);
