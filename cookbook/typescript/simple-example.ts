#!/usr/bin/env bun

import { Retab, Schema } from '@retab/node';
import * as process from 'process';

async function main() {
  console.log('✅ TypeScript SDK Example');
  console.log('========================\n');

  // Demonstrate type safety with TypeScript
  const createClient = (apiKey: string): Retab => {
    return new Retab({ apiKey });
  };

  // Example schema with full type safety
  const invoiceSchema = Schema.from_json_schema({
    type: 'object',
    properties: {
      invoice_number: { type: 'string' },
      total_amount: { type: 'number' },
      line_items: {
        type: 'array',
        items: {
          type: 'object',
          properties: {
            description: { type: 'string' },
            amount: { type: 'number' }
          }
        }
      }
    },
    required: ['invoice_number', 'total_amount']
  });

  console.log('📄 Schema created successfully');
  console.log('Schema ID:', invoiceSchema.id);
  console.log('Schema type:', invoiceSchema.type);

  // Demonstrate async function with proper types
  async function extractExample() {
    if (!process.env.RETAB_API_KEY) {
      console.log('\n⚠️  Note: Set RETAB_API_KEY to run actual extractions');
      return;
    }
    
    const client = createClient(process.env.RETAB_API_KEY);
    console.log('\n🚀 Ready to extract documents!');
    
    // Example of typed extraction
    /*
    const result = await client.documents.extract({
      document: 'path/to/invoice.pdf',
      schema: invoiceSchema,
      model: 'gpt-4o'
    });
    
    console.log('Extracted data:', result.extracted_data);
    */
  }

  // Run the example
  await extractExample();
  console.log('\n✨ TypeScript example complete!');
}

main().catch(console.error);