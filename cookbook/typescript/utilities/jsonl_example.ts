import * as fs from 'fs';
import { 
  readJSONL, 
  writeJSONL, 
  appendJSONL, 
  filterJSONL, 
  splitJSONL, 
  validateJSONL, 
  streamJSONL 
} from '@retab/node/dist/utils/jsonl.js';
import { config } from 'dotenv';

async function main() {
  // ---------------------------------------------
  // Example: JSONL Utilities Usage
  // ---------------------------------------------

  // Load environment variables
  config();

  async function demonstrateJSONLUtils() {
    console.log('=== JSONL Utilities Demo ===\n');
    
    const testData = [
      { id: 1, name: 'Alice', age: 30, city: 'New York' },
      { id: 2, name: 'Bob', age: 25, city: 'San Francisco' },
      { id: 3, name: 'Charlie', age: 35, city: 'Chicago' },
      { id: 4, name: 'Diana', age: 28, city: 'Boston' },
      { id: 5, name: 'Eve', age: 32, city: 'Seattle' },
    ];
    
    const testFile = 'test_data.jsonl';
    const filteredFile = 'filtered_data.jsonl';
    
    try {
      // Write data to JSONL file
      console.log('1. Writing data to JSONL file...');
      await writeJSONL(testFile, testData);
      console.log(`✅ Written ${testData.length} records to ${testFile}`);
      
      // Read data back
      console.log('\n2. Reading data from JSONL file...');
      const readData = await readJSONL(testFile);
      console.log(`✅ Read ${readData.length} records:`);
      console.log(readData.slice(0, 2)); // Show first 2 records
      
      // Append new data
      console.log('\n3. Appending new record...');
      const newRecord = { id: 6, name: 'Frank', age: 29, city: 'Denver' };
      await appendJSONL(testFile, newRecord);
      console.log('✅ Appended new record');
      
      // Validate file
      console.log('\n4. Validating JSONL file...');
      const validation = await validateJSONL(testFile);
      console.log(`✅ Validation result: ${validation.valid ? 'Valid' : 'Invalid'}`);
      if (!validation.valid) {
        console.log('Errors:', validation.errors);
      }
      
      // Filter data
      console.log('\n5. Filtering data (age >= 30)...');
      const filteredCount = await filterJSONL(
        testFile,
        filteredFile,
        (record) => record.age >= 30
      );
      console.log(`✅ Filtered ${filteredCount} records to ${filteredFile}`);
      
      // Read filtered data
      const filteredData = await readJSONL(filteredFile);
      console.log('Filtered records:', filteredData);
      
      // Split into chunks
      console.log('\n6. Splitting into chunks (2 records each)...');
      const chunkFiles = await splitJSONL(testFile, './chunks', 2);
      console.log(`✅ Created ${chunkFiles.length} chunk files:`);
      chunkFiles.forEach(file => console.log(`  - ${file}`));
      
      // Demonstrate streaming processing
      console.log('\n7. Streaming transformation...');
      const transformedFile = 'transformed_data.jsonl';
      await streamJSONL(
        testFile,
        transformedFile,
        (record) => ({
          ...record,
          age_group: record.age < 30 ? 'young' : 'mature',
          name_length: record.name.length,
        })
      );
      
      const transformedData = await readJSONL(transformedFile);
      console.log('✅ Transformed data sample:');
      console.log(transformedData.slice(0, 2));
      
      // Clean up
      [testFile, filteredFile, transformedFile, ...chunkFiles].forEach(file => {
        if (fs.existsSync(file)) {
          fs.unlinkSync(file);
        }
      });
      
      // Clean up chunks directory
      if (fs.existsSync('./chunks')) {
        fs.rmSync('./chunks', { recursive: true });
      }
      
      console.log('\n✅ JSONL utilities demo completed successfully!');
      
    } catch (error) {
      console.error('Error in JSONL demo:', error);
    }
  }

  async function demonstrateRealWorldUse() {
    console.log('\n=== Real-World JSONL Example ===\n');
    
    // Simulate training data for ML model
    const trainingData = [
      { input: 'Hello world', output: 'greeting', confidence: 0.95 },
      { input: 'How are you?', output: 'question', confidence: 0.88 },
      { input: 'Thank you', output: 'gratitude', confidence: 0.92 },
      { input: 'Goodbye', output: 'farewell', confidence: 0.87 },
    ];
    
    const trainingFile = 'training_data.jsonl';
    const highConfidenceFile = 'high_confidence.jsonl';
    
    try {
      // Write training data
      await writeJSONL(trainingFile, trainingData);
      console.log('✅ Created training data file');
      
      // Filter high confidence samples
      const highConfCount = await filterJSONL(
        trainingFile,
        highConfidenceFile,
        (sample) => sample.confidence >= 0.9
      );
      
      console.log(`✅ Found ${highConfCount} high confidence samples`);
      
      const highConfData = await readJSONL(highConfidenceFile);
      console.log('High confidence samples:', highConfData);
      
      // Clean up
      [trainingFile, highConfidenceFile].forEach(file => {
        if (fs.existsSync(file)) {
          fs.unlinkSync(file);
        }
      });
      
      console.log('\n✅ Real-world example completed!');
      
    } catch (error) {
      console.error('Error in real-world example:', error);
    }
  }

  // Run the demos
  async function runDemo() {
    await demonstrateJSONLUtils();
    await demonstrateRealWorldUse();
  }

  await runDemo();
}

main().catch(console.error);