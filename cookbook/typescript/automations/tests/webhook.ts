import { config } from 'dotenv';
import { Retab } from '@retab/node';



async function main() {
  // Load environment variables
  config();
  
  const client = new Retab();
  
  const link = await client.processors.automations.links.create({
    name: "Link Automation 4",
    processor_id: "proc_2BTqDkKOTO_Ddz1N7JwVZ",
    webhook_url: "http://localhost:4000/test-webhook",
  });
  
  // If you just want to send a test request to your webhook
  const log = await client.processors.automations.tests.webhook({
    automation_id: link.id,
    completion: { /* completion object */ },
    file_payload: { /* file payload */ },
  });
  
  console.log(JSON.stringify(log, null, 2));
  
}

main().catch(console.error);