import { config } from 'dotenv';
import { Retab } from '@retab/node';

// Load environment variables
config({ path: '../../../.env.local' });

const client = new Retab();

// If you want to test the file processing logic:
const log = await client.processors.automations.tests.upload({
  automation_id: "lnk_DwHVXFUVLgIkxXQaxvx2s",
  document: "your-invoice-email.eml"
});

console.log(JSON.stringify(log, null, 2));