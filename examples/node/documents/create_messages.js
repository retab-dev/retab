import OpenAI from 'openai';
import { Retab } from '@retab/node';
import { config } from 'dotenv';

// ---------------------------------------------
// Quick example: Summarize a document using Retab + OpenAI, without a schema.
// ---------------------------------------------

// Load environment variables
config();

const apiKey = process.env.OPENAI_API_KEY;
const retabApiKey = process.env.RETAB_API_KEY;

if (!apiKey) {
  throw new Error('Missing OPENAI_API_KEY');
}
if (!retabApiKey) {
  throw new Error('Missing RETAB_API_KEY');
}

// Retab Setup
const reclient = new Retab({ api_key: retabApiKey });
const docMsg = await reclient.documents.create_messages({
  document: '../../assets/code/booking_confirmation.jpg',
  // Note: Additional parameters like modality, image_resolution_dpi, browser_canvas 
  // would be supported by extending the create_messages implementation
});

const client = new OpenAI({ apiKey });
const completion = await client.chat.completions.create({
  model: 'gpt-4o-mini',
  messages: [
    ...docMsg.openai_messages,
    { role: 'user', content: 'Summarize the document' }
  ]
});

// Output
console.log('\n📝 OpenAI Completion Output:');
console.log(JSON.stringify(completion, null, 2));