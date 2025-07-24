
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
