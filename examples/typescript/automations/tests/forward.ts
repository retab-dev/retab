import * as process from 'process';
import { config } from 'dotenv';
import { Retab } from '@retab/node';



async function main() {
  config({ path: '../../../.env.production' });
  
  const client = new Retab();
  
  const mailbox = await client.processors.automations.mailboxes.create({
    name: "Mailbox Automation",
    processor_id: "proc_2BTqDkKOTO_Ddz1N7JwVZ",
    webhook_url: "http://api.retab.com/webhook",
    webhook_headers: { "API-Key": process.env.RETAB_API_KEY },
    email: "invoices-4@mailbox.retab.com",
  });
  
  console.log(JSON.stringify(mailbox, null, 2));
  
  // If you want to test a full email forwarding
  const log = await client.processors.automations.mailboxes.tests.forward({
    email: "invoices-3@mailbox.retab.com", 
    document: "your-invoice-email.eml"
  });
  
  console.log(JSON.stringify(log, null, 2));
  
}

main().catch(console.error);