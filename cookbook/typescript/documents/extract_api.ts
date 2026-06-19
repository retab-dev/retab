import * as process from 'process';
import * as fs from 'fs';
import { Retab } from '@retab/node';
import { config } from 'dotenv';


async function main() {
  // ---------------------------------------------
  // Quick example: Extract structured data using `client.extractions.create()`.
  // ---------------------------------------------


  // Load environment variables
  config();

  const retabApiKey = process.env.RETAB_API_KEY;
  if (!retabApiKey) {
    throw new Error('Missing RETAB_API_KEY');
  }

  // Retab Setup
  const client = new Retab({ apiKey: retabApiKey });

  // Document Extraction via Retab API.
  // create() takes positional args: (document, jsonSchema, model?, instructions?, nConsensus?)
  const response = await client.extractions.create(
    '../../../assets/docs/invoice.jpeg',
    JSON.parse(fs.readFileSync('../../../assets/code/invoice_schema.json', 'utf-8')),
    'retab-small',
    undefined, // instructions
    1, // nConsensus
  );

  // Output
  console.log('\n✅ Extracted Result:');
  console.log(response.output);

}

main().catch(console.error);
