#!/usr/bin/env bun

import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log('🎯 Final TypeScript Examples Test\n');

// Test 1: Basic TypeScript compilation
console.log('1. Testing TypeScript compilation...');
try {
  execSync('npx tsc --noEmit', { 
    cwd: __dirname,
    stdio: 'pipe'
  });
  console.log('✅ TypeScript compilation successful\n');
} catch (error) {
  console.log('❌ TypeScript compilation failed');
  console.log('Error output:', error.stdout?.toString() || error.stderr?.toString());
  console.log('');
}

// Test 2: SDK Import test
console.log('2. Testing SDK imports...');
try {
  const testImport = `
import { Retab, Schema } from '@retab/node';
import { 
  readJSONL, 
  writeJSONL 
} from '@retab/node/dist/utils/jsonl.js';
import { 
  optimizePrompt, 
  calculateMetrics 
} from '@retab/node/dist/utils/prompt_optimization.js';

console.log('✅ All SDK imports successful');
`;
  
  fs.writeFileSync(path.join(__dirname, 'temp-import-test.ts'), testImport);
  execSync('bun run temp-import-test.ts', {
    cwd: __dirname,
    stdio: 'pipe'
  });
  fs.unlinkSync(path.join(__dirname, 'temp-import-test.ts'));
  console.log('✅ SDK imports working correctly\n');
} catch (error) {
  console.log('❌ SDK import test failed');
  console.log('Error:', error.message);
  console.log('');
}

// Test 3: Basic client creation
console.log('3. Testing client creation...');
try {
  const testClient = `
import { Retab, Schema } from '@retab/node';

// Test client creation
const client = new Retab({ apiKey: 'test-key' });
console.log('✅ Client created successfully');

// Test schema creation
const schema = Schema.from_json_schema({
  type: 'object',
  properties: {
    name: { type: 'string' },
    age: { type: 'number' }
  }
});
console.log('✅ Schema created successfully');
`;
  
  fs.writeFileSync(path.join(__dirname, 'temp-client-test.ts'), testClient);
  execSync('bun run temp-client-test.ts', {
    cwd: __dirname,
    stdio: 'pipe'
  });
  fs.unlinkSync(path.join(__dirname, 'temp-client-test.ts'));
  console.log('✅ Client creation working correctly\n');
} catch (error) {
  console.log('❌ Client creation test failed');
  console.log('Error:', error.message);
  console.log('');
}

// Test 4: Count converted files
console.log('4. Counting converted examples...');
const tsFiles = [];
function findTsFiles(dir) {
  const items = fs.readdirSync(dir);
  for (const item of items) {
    const fullPath = path.join(dir, item);
    const stat = fs.statSync(fullPath);
    if (stat.isDirectory() && item !== 'node_modules') {
      findTsFiles(fullPath);
    } else if (item.endsWith('.ts') && !item.startsWith('test-') && !item.startsWith('temp-')) {
      tsFiles.push(fullPath);
    }
  }
}

findTsFiles(__dirname);
console.log(`✅ Found ${tsFiles.length} TypeScript examples\n`);

// Test 5: Structure validation
console.log('5. Validating directory structure...');
const expectedDirs = [
  'automations',
  'consensus', 
  'documents',
  'processors',
  'schemas',
  'utilities'
];

let structureValid = true;
for (const dir of expectedDirs) {
  const dirPath = path.join(__dirname, dir);
  if (!fs.existsSync(dirPath)) {
    console.log(`❌ Missing directory: ${dir}`);
    structureValid = false;
  }
}

if (structureValid) {
  console.log('✅ Directory structure valid\n');
} else {
  console.log('❌ Directory structure issues found\n');
}

console.log('🎉 Final test complete!');
console.log('\nTo run examples:');
console.log('  bun run simple-example.ts');
console.log('  bun run utilities/jsonl_example.ts');
console.log('  bun run utilities/prompt_optimization_example.ts');
console.log('\nNote: Most examples require API keys to be set in .env file');