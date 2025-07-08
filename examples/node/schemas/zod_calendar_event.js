// ---------------------------------------------
// Example: Define and use a CalendarEvent schema using Zod (recommended for Node.js/TypeScript devs)
// ---------------------------------------------

import OpenAI from 'openai';
import { config } from 'dotenv';
import { z } from 'zod';
import { Schema, Retab } from '@retab/node';
import { filterAuxiliaryFieldsJson } from '@retab/node/utils/json_schema';

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

const reclient = new Retab({ api_key: retabApiKey });
const docMsg = await reclient.documents.create_messages({ document: '../../../assets/code/calendar_event.xlsx' });

console.log('Document message:', JSON.stringify(docMsg, null, 2));

// Define schema using Zod
const CalendarEventSchema = z.object({
  name: z.string().describe('The name of the calendar event.'),
  date: z.string().describe('The date of the calendar event in ISO 8601 format.'),
});

// Create schema object with Zod model
const schemaObj = new Schema({
  zod_model: CalendarEventSchema,
  system_prompt: 'You are a useful assistant extracting information from documents.',
  reasoning_prompts: {
    date: 'The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.',
  },
});

// Transform the messages to OpenAI format
const openaiMessages = docMsg.messages || [];

// Now you can use your favorite model to analyze your document
const client = new OpenAI({ apiKey });
const completion = await client.chat.completions.create({
  model: 'gpt-4o',
  messages: [...schemaObj.openai_messages, ...openaiMessages],
  response_format: { 
    type: 'json_schema', 
    json_schema: { 
      name: schemaObj.id, 
      schema: schemaObj.inference_json_schema, 
      strict: true 
    } 
  },
});


// Debug the completion response
console.log('Completion response:', JSON.stringify(completion, null, 2));

// Validate the response against the original schema if you want to remove the reasoning fields
if (!completion.choices[0].message.content) {
  throw new Error('No content in response');
}

const extraction = schemaObj.zod_model.parse(filterAuxiliaryFieldsJson(completion.choices[0].message.content));

console.log('\n✅ Extracted Calendar Event:');
console.log('Extraction:', extraction);