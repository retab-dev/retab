import * as process from 'process';
import * as fs from 'fs';
import { Retab } from '@retab/node';
import { config } from 'dotenv';


async function main() {
  // ---------------------------------------------
  // Quick example: Extract structured data using Retab's all-in-one `.parse()` method.
  // ---------------------------------------------


  // Load environment variables
  config();

  const retabApiKey = process.env.RETAB_API_KEY;
  if (!retabApiKey) {
    throw new Error('Missing RETAB_API_KEY');
  }

  // Retab Setup
  const client = new Retab({ apiKey: retabApiKey });

  // Document Extraction via Retab API
  const response = await client.documents.extract({
    documents: ['../../assets/docs/invoice.jpeg'],
    model: 'gpt-5',
    json_schema: JSON.parse(fs.readFileSync('../../assets/code/invoice_schema.json', 'utf-8')),
    temperature: 0.0,
    n_consensus: 1,
  });

  // Output
  console.log('\n✅ Extracted Result:');
  console.log(response.choices[0].message.content);

}

main().catch(console.error);