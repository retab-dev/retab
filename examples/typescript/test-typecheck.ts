#!/usr/bin/env bun

import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log('🔍 Running TypeScript type checking on all examples...\n');

// Find all TypeScript files
function findTsFiles(dir: string, files: string[] = []): string[] {
  const items = fs.readdirSync(dir);
  
  for (const item of items) {
    const fullPath = path.join(dir, item);
    const stat = fs.statSync(fullPath);
    
    if (stat.isDirectory() && item !== 'node_modules') {
      findTsFiles(fullPath, files);
    } else if (item.endsWith('.ts') && item !== 'test-typecheck.ts') {
      files.push(fullPath);
    }
  }
  
  return files;
}

const tsFiles = findTsFiles(__dirname);
console.log(`Found ${tsFiles.length} TypeScript files to check.\n`);

// Run tsc on each file
let errors = 0;
let success = 0;

for (const file of tsFiles) {
  const relativePath = path.relative(__dirname, file);
  try {
    execSync(`npx tsc --noEmit ${file}`, { 
      cwd: __dirname,
      stdio: 'pipe'
    });
    console.log(`✅ ${relativePath}`);
    success++;
  } catch (error) {
    console.log(`❌ ${relativePath}`);
    if (error instanceof Error && 'stdout' in error) {
      console.log(error.stdout?.toString());
    }
    errors++;
  }
}

console.log(`\n📊 Summary:`);
console.log(`   ✅ Success: ${success}`);
console.log(`   ❌ Errors: ${errors}`);
console.log(`   📁 Total: ${tsFiles.length}`);

if (errors > 0) {
  console.log('\n⚠️  Some files have type errors. Running full typecheck...\n');
  try {
    execSync('npx tsc --noEmit', { cwd: __dirname, stdio: 'inherit' });
  } catch (e) {
    // Type errors will be shown above
  }
}

process.exit(errors > 0 ? 1 : 0);