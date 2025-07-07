import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const client = new AsyncRetab();

console.log("Note: Automations sub-resources are not yet fully implemented in the Node.js SDK");
console.log("This is a placeholder example showing the intended structure");

// This would be the intended usage once automations are fully implemented:
/*
const mailbox = await client.processors.automations.mailboxes.create({
  name: "Invoice Mailbox",
  email: "invoices@mailbox.retab.com",
  processor_id: "proc_01G34H8J2K",  // The processor id you created in the previous step
  webhook_url: "https://your-server.com/webhook",  // Replace with your actual webhook URL
});

// If you just want to send a test request to your webhook
const log = await client.processors.automations.tests.webhook({
  automation_id: mailbox.id,
});

// If you want to test the file processing logic:
const log2 = await client.processors.automations.tests.upload({ 
  automation_id: mailbox.id, 
  document: "your_invoice_email.eml" 
});

// If you want to test a full email forwarding
const log3 = await client.processors.automations.mailboxes.tests.forward({ 
  email: "invoices@mailbox.retab.com", 
  document: "your_invoice_email.eml" 
});
*/