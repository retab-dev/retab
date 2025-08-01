// ---------------------------------------------
// Full example: Use Retab with OpenAI's Responses API and JSON Schema extraction.
// ---------------------------------------------

import OpenAI from 'openai';
import { Schema, Retab } from '@retab/node';
import { filterAuxiliaryFieldsJson } from '@retab/node/utils/json_schema';
import { config } from 'dotenv';

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

// Define schema
const jsonSchema = {
  'X-SystemPrompt': 'You are a useful assistant extracting information from documents.',
  properties: {
    name: {
      description: 'The name of the calendar event.',
      title: 'Name',
      type: 'string',
    },
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
};

// Configuration
const model = 'gpt-4o';
const modality = 'native';
const temperature = 0.0;
const imageResolutionDpi = 96;
const browserCanvas = 'A4';

// Retab Setup
const reclient = new Retab({ api_key: retabApiKey });
const docMsg = await reclient.documents.create_messages({
  document: '../../assets/docs/calendar_event.xlsx',
  // Note: Additional parameters like modality, image_resolution_dpi, browser_canvas 
  // would be supported by extending the create_messages implementation
});
const schemaObj = new Schema({ json_schema: jsonSchema });

// Transform the messages to OpenAI format
const openaiMessages = docMsg.openai_messages || [];

// OpenAI Responses API call
const client = new OpenAI({ apiKey });

// Note: The OpenAI Node.js SDK may not support the responses API in the same way as Python
// This is a conceptual example showing how it would work:

try {
  // Example 1: With hypothetical client.responses.create() method
  console.log('Note: OpenAI Responses API may not be available in Node.js SDK');
  console.log('Using regular completions API instead...');
  
  const response = await client.chat.completions.create({
    model,
    temperature,
    messages: [...schemaObj.openai_messages, ...openaiMessages],
    response_format: {
      type: 'json_schema',
      json_schema: {
        name: schemaObj.id,
        schema: schemaObj.inference_json_schema,
        strict: true,
      },
    },
  });

  // Validate response and remove reasoning fields
  if (!response.choices[0].message.content) {
    throw new Error('No content in response');
  }

  const parsedContent = JSON.parse(response.choices[0].message.content);
  const extraction = filterAuxiliaryFieldsJson(parsedContent);

  // Output
  console.log('\n✅ Extracted Result (Completions API):');
  console.log(JSON.stringify(extraction, null, 2));

} catch (error) {
  console.error('Error with API call:', error.message);
  console.log('Note: This example demonstrates the structure for OpenAI Responses API integration');
}