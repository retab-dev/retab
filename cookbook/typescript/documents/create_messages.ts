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
  const result = await client.documents.parse({
    document: '../../assets/docs/booking_confirmation.jpg',
    // Note: Additional parameters like image_resolution_dpi, browser_canvas 
  });

  console.log(result.pages);


}

main().catch(console.error);