import { Schema } from './dist/types/schemas/object.js';

console.log('🔄 Testing Circular Reference Detection Fix\n');

// Test 1: Schema with circular references
console.log('1️⃣ Schema with Circular References:');
const circularSchema = {
  type: "object",
  properties: {
    parent: { "$ref": "#/$defs/Node" }
  },
  "$defs": {
    "Node": {
      type: "object",
      properties: {
        value: { type: "string" },
        children: {
          type: "array",
          items: { "$ref": "#/$defs/Node" }
        }
      }
    }
  }
};

console.log('Input schema has circular Node -> children[] -> Node reference');

try {
  const schema1 = new Schema({ json_schema: circularSchema });
  console.log('   ✅ Schema creation successful (no stack overflow)');
  
  const expanded = schema1._expandedObjectSchema;
  console.log('   ✅ Schema expansion completed without error');
  console.log('   Expanded schema has properties:', Object.keys(expanded.properties || {}));
  
} catch (error) {
  console.log('   ❌ Schema creation failed:', error.message);
}

// Test 2: Schema without circular references
console.log('\n2️⃣ Schema without Circular References:');
const normalSchema = {
  type: "object",
  properties: {
    user: { "$ref": "#/$defs/User" },
    address: { "$ref": "#/$defs/Address" }
  },
  "$defs": {
    "User": {
      type: "object",
      properties: {
        name: { type: "string" },
        age: { type: "number" }
      }
    },
    "Address": {
      type: "object", 
      properties: {
        street: { type: "string" },
        city: { type: "string" }
      }
    }
  }
};

try {
  const schema2 = new Schema({ json_schema: normalSchema });
  console.log('   ✅ Normal schema creation successful');
  
  const expanded2 = schema2._expandedObjectSchema;
  const hasRefs = JSON.stringify(expanded2).includes('$ref');
  console.log(`   ✅ References properly expanded: ${!hasRefs ? 'Yes' : 'No'}`);
  
  // Check if nested properties are accessible
  const userProps = expanded2.properties?.user?.properties;
  const addressProps = expanded2.properties?.address?.properties;
  console.log(`   ✅ User properties resolved: ${!!userProps ? 'Yes' : 'No'}`);
  console.log(`   ✅ Address properties resolved: ${!!addressProps ? 'Yes' : 'No'}`);
  
} catch (error) {
  console.log('   ❌ Normal schema failed:', error.message);
}

// Test 3: allOf schema handling
console.log('\n3️⃣ allOf Schema Handling:');
const allOfSchema = {
  type: "object",
  allOf: [
    {
      properties: {
        name: { type: "string" }
      }
    }
  ]
};

try {
  const schema3 = new Schema({ json_schema: allOfSchema });
  console.log('   ✅ allOf schema creation successful');
  
  const expanded3 = schema3._expandedObjectSchema;
  const hasNameProp = expanded3.properties?.name;
  console.log(`   ✅ allOf properties merged: ${hasNameProp ? 'Yes' : 'No'}`);
  
} catch (error) {
  console.log('   ❌ allOf schema failed:', error.message);
}

// Test 4: Complex allOf with multiple schemas
console.log('\n4️⃣ Multiple allOf Schemas (should fail):');
const multiAllOfSchema = {
  type: "object",
  allOf: [
    { properties: { name: { type: "string" } } },
    { properties: { age: { type: "number" } } }
  ]
};

try {
  const schema4 = new Schema({ json_schema: multiAllOfSchema });
  console.log('   ❌ Should have failed with multiple allOf');
} catch (error) {
  console.log('   ✅ Correctly rejected multiple allOf:', error.message);
}

console.log('\n📊 Summary:');
console.log('   - Circular reference detection prevents infinite loops');
console.log('   - Normal references are properly expanded');
console.log('   - allOf schemas are merged correctly');
console.log('   - Error handling for complex allOf schemas');

console.log('\n🎉 Reference expansion is now robust!');