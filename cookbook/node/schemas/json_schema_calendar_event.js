// ---------------------------------------------
// Example: Define and use a CalendarEvent schema using JSON Schema (manual)
// ---------------------------------------------

import OpenAI from 'openai';
import { config } from 'dotenv';
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
const docMsg = await reclient.documents.create_messages({ document: '../../assets/docs/calendar_event.xlsx' });

console.log('Document message:', JSON.stringify(docMsg, null, 2));

// Define schema
const schemaObj = new Schema({
  json_schema: {
    'X-SystemPrompt': 'You are a useful assistant extracting information from documents.',
    properties: {
      name: { description: 'The name of the calendar event.', title: 'Name', type: 'string' },
      date: {
        'X-ReasoningPrompt': 'The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.',
        description: 'The date of the calendar event in ISO 8601 format.',
        title: 'Date',
        type: 'string',
      },
    },
    required: ['name', 'date'],
    title: 'CalendarEvent',
    type: 'object',
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

// For JSON schema, we just parse and filter the auxiliary fields
const extraction = filterAuxiliaryFieldsJson(completion.choices[0].message.content);

console.log('\n✅ Extracted Calendar Event:');
console.log(JSON.stringify(extraction, null, 2));