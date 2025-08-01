// ---------------------------------------------
// Quick example: Extract structured data using Retab's all-in-one `.parse()` method from buffer.
// ---------------------------------------------

import { Retab } from '@retab/node';
import { config } from 'dotenv';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

// Load environment variables
config();

const retabApiKey = process.env.RETAB_API_KEY;
if (!retabApiKey) {
  throw new Error('Missing RETAB_API_KEY');
}

// Get current directory for relative path resolution
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Read the invoice image into a buffer
const imagePath = join(__dirname, '../../../assets/docs/invoice.jpeg');
const imageBuffer = readFileSync(imagePath);

// Read the JSON schema
const schemaPath = join(__dirname, '../../../assets/code/invoice_schema.json');
const jsonSchema = readFileSync(schemaPath, 'utf8');

// Retab Setup
const reclient = new Retab({ api_key: retabApiKey });

// Document Extraction via Retab API using buffer
const response = await reclient.documents.extract({
  document: imageBuffer,
  model: 'gpt-4.1',
  json_schema: jsonSchema,
  modality: 'native',
  image_resolution_dpi: 96,
  browser_canvas: 'A4',
  temperature: 0.0,
  n_consensus: 1,
});

// Output
console.log('\n✅ Extracted Result:');
console.log(response.choices[0].message.content);