import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const client = new AsyncRetab();

const mailbox = await client.processors.automations.mailboxes.create({
  name: "Invoice Mailbox",
  email: "invoices@mailbox.retab.com",
  processor_id: "proc_01G34H8J2K",  // The processor id you created in the previous step
  webhook_url: "https://your-server.com/webhook",  // Replace with your actual webhook URL
});

console.log('Created mailbox:', mailbox);

// If you just want to send a test request to your webhook
const log = await client.processors.automations.tests.webhook({
  automation_id: mailbox.id,
  completion: { /* completion object */ },
  file_payload: { /* file payload */ },
});

console.log('Webhook test result:', log);

// If you want to test the file processing logic:
const log2 = await client.processors.automations.tests.upload({ 
  automation_id: mailbox.id, 
  document: "your_invoice_email.eml" 
});

console.log('Upload test result:', log2);

// If you want to test a full email forwarding
const log3 = await client.processors.automations.tests.forward({ 
  automation_id: mailbox.id,
  email_file: "your_invoice_email.eml" 
});

console.log('Forward test result:', log3);