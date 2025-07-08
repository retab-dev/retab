// ---------------------------------------------
// Full example: Extract structured data using Retab + OpenAI Chat Completion + JSON Schema
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

// Define model and modality
const model = 'gpt-4o';
const modality = 'native';
const temperature = 0.0;
const imageResolutionDpi = 96;
const browserCanvas = 'A4';

// Retab Setup
const reclient = new Retab({ api_key: retabApiKey });
const docMsg = await reclient.documents.create_messages({
  document: '../../assets/code/calendar_event.xlsx',
  // Note: Additional parameters like modality, image_resolution_dpi, browser_canvas 
  // would be supported by extending the create_messages implementation
});
const schemaObj = new Schema({ json_schema: jsonSchema });

// Transform the messages to OpenAI format
const openaiMessages = docMsg.openai_messages || [];

// OpenAI Chat Completion with schema-based prompting
const client = new OpenAI({ apiKey });
// Example 1: Use the `inference_json_schema` parameter
const completion = await client.chat.completions.create({
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

// Note: Example 2 with beta.chat.completions.parse would be:
// const completion = await client.beta.chat.completions.parse({
//   model,
//   messages: [...schemaObj.openai_messages, ...openaiMessages],
//   response_format: schemaObj.inference_pydantic_model,
// });

// Validate and clean the output
if (!completion.choices[0].message.content) {
  throw new Error('No content in response');
}

const extraction = schemaObj.pydantic_model.model_validate(filterAuxiliaryFieldsJson(completion.choices[0].message.content));

// Output
console.log('\n✅ Extracted Result:');
console.log(JSON.stringify(extraction, null, 2));