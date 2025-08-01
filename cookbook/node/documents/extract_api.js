// ---------------------------------------------
// Quick example: Extract structured data using Retab's all-in-one `.parse()` method.
// ---------------------------------------------

import { Retab } from '@retab/node';
import { config } from 'dotenv';

// Load environment variables
config();

const retabApiKey = process.env.RETAB_API_KEY;
if (!retabApiKey) {
  throw new Error('Missing RETAB_API_KEY');
}

// Retab Setup
const reclient = new Retab({ api_key: retabApiKey });

// Document Extraction via Retab API
const response = await reclient.documents.extract({
  document: '../../assets/docs/invoice.jpeg',
  model: 'gpt-4.1',
  json_schema: '../../assets/code/invoice_schema.json',
  modality: 'native',
  image_resolution_dpi: 96,
  browser_canvas: 'A4',
  temperature: 0.0,
  n_consensus: 1,
});

// Output
console.log('\n✅ Extracted Result:');
console.log(response.choices[0].message.content);