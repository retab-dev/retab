import * as process from 'process';
import { Retab } from '@retab/node';
import { config } from 'dotenv';

async function main() {
  // ---------------------------------------------
  // Quick example: Parse a document in markdown format using Retab.
  // ---------------------------------------------

  // Load environment variables
  config();

  const retabApiKey = process.env.RETAB_API_KEY;

  if (!retabApiKey) {
    throw new Error('Missing RETAB_API_KEY');
  }

  // Retab Setup
  const client = new Retab({ apiKey: retabApiKey });
  // create() takes positional args: (document, model?, tableParsingFormat?, imageResolutionDpi?)
  const result = await client.parses.create(
    '../../../assets/docs/booking_confirmation.jpg',
    'retab-small',
    undefined, // tableParsingFormat
    192, // imageResolutionDpi
  );

  console.log(result.output.pages);


}

main().catch(console.error);
