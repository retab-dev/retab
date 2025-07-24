import { config } from 'dotenv';
import { Retab } from '@retab/node';

config({ path: '../../../.env.local' });

const client = new Retab();

const link = await client.processors.automations.links.create({
  name: "Link Automation",
  processor_id: "proc_o4dtLxizT0kDAjeKuyVLA",
  webhook_url: "https://api.retab.com/webhook",
  webhook_headers: { "API-Key": process.env.RETAB_API_KEY },
});

console.log(JSON.stringify(link, null, 2));

// If you want to test the file processing logic:
const log = await client.processors.automations.tests.upload({
  automation_id: link.id, 
  document: "your-invoice-email.eml"
});

console.log(JSON.stringify(log, null, 2));