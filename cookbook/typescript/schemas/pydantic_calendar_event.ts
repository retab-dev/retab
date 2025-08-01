import * as process from 'process';
import OpenAI from 'openai';
import { config } from 'dotenv';
import { Schema, Retab } from '@retab/node';
import { filterAuxiliaryFieldsJson } from '@retab/node/dist/utils/json_schema.js';


async function main() {
  // ---------------------------------------------
  // Example: Define and use a CalendarEvent schema using a Pydantic-like approach (equivalent to Python Pydantic)
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
  
  const reclient = new Retab({ apiKey: retabApiKey });
  const docMsg = await reclient.documents.create_messages({ document: '../../assets/docs/calendar_event.xlsx' });
  
  // Define schema using JSON Schema (equivalent to Pydantic BaseModel)
  const CalendarEventJsonSchema = {
    "X-SystemPrompt": "You are a useful assistant.",
    "properties": {
      "name": {
        "description": "The name of the calendar event.",
        "title": "Name",
        "type": "string"
      },
      "date": {
        "X-ReasoningPrompt": "The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.",
        "description": "The date of the calendar event in ISO 8601 format.",
        "title": "Date",
        "type": "string"
      }
    },
    "required": ["name", "date"],
    "title": "CalendarEvent",
    "type": "object"
  };
  
  const schemaObj = new Schema({ json_schema: CalendarEventJsonSchema });
  
  // Now you can use your favorite model to analyze your document
  const client = new OpenAI({ apiKey });
  
  // Use the beta.chat.completions.parse equivalent (structured outputs)
  const completion = await client.chat.completions.create({
    model: 'gpt-4o',
    messages: [...schemaObj.openai_messages, ...docMsg.messages],
    response_format: { 
      type: 'json_schema', 
      json_schema: { 
        name: schemaObj.id, 
        schema: schemaObj.inference_json_schema, 
        strict: true 
      } 
    },
  });
  
  // Debug the completion
  console.log('Completion:', JSON.stringify(completion, null, 2));
  
  // Validate the response against the original schema if you want to remove the reasoning fields
  if (!completion.choices[0].message.content) {
    throw new Error('No content in response');
  }
  
  // Parse and validate using JSON parse and filter
  const rawContent = completion.choices[0].message.content;
  const parsedContent = JSON.parse(rawContent);
  const extraction = filterAuxiliaryFieldsJson(parsedContent);
  
  console.log('\n✅ Extracted Calendar Event:');
  console.log('Extraction:', extraction);
  
}

main().catch(console.error);