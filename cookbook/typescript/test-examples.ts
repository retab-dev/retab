#!/usr/bin/env bun

import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log('🧪 Testing TypeScript examples with Bun...\n');

// Examples that don't require API keys or external resources
const testableExamples = [
  'schemas/json_schema_calendar_event.ts',
  'utilities/prompt_optimization_example.ts',
];

console.log('Testing syntax by running examples that don\'t require API keys:\n');

for (const example of testableExamples) {
  const fullPath = path.join(__dirname, example);
  console.log(`📄 Testing ${example}...`);
  
  try {
    // Check if file exists
    if (!fs.existsSync(fullPath)) {
      console.log(`   ❌ File not found\n`);
      continue;
    }
    
    // Try to run with --dry-run to check syntax
    execSync(`bun run --dry ${fullPath}`, {
      cwd: __dirname,
      stdio: 'pipe'
    });
    console.log(`   ✅ Syntax OK\n`);
  } catch (error) {
    console.log(`   ❌ Syntax error\n`);
    if (error instanceof Error && 'stderr' in error) {
      console.log(error.stderr?.toString());
    }
  }
}

// Test that the SDK can be imported
console.log('📦 Testing SDK import...');
try {
  const testImport = `
import { Retab, Schema } from '@retab/node';
console.log('SDK imported successfully');
`;
  
  fs.writeFileSync(path.join(__dirname, 'test-import.ts'), testImport);
  execSync('bun run test-import.ts', {
    cwd: __dirname,
    stdio: 'pipe'
  });
  fs.unlinkSync(path.join(__dirname, 'test-import.ts'));
  console.log('   ✅ SDK imports correctly\n');
} catch (error) {
  console.log('   ❌ SDK import failed\n');
}

console.log('✨ Test complete!');
console.log('\nTo run any example with your API keys:');
console.log('1. Create a .env file with RETAB_API_KEY and other required keys');
console.log('2. Run: bun run <example-path>');
console.log('\nExample: bun run documents/extract_api.ts');