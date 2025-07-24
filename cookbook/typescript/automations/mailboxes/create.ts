import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';



async function main() {
  config();
  
  const reclient = new AsyncRetab();
  
  console.log("Note: Mailbox automations are not yet fully implemented in the Node.js SDK");
  console.log("This is a placeholder example showing the intended structure");
  
  // This would be the intended usage once mailbox automations are fully implemented:
  /*
  const automation = await reclient.processors.automations.mailboxes.create({
    name: "Mailbox Automation",
    processor_id: "proc_2BTqDkKOTO_Ddz1N7JwVZ",
    webhook_url: "https://api.retab.com/test-webhook",
    email: "invoices-new@mailbox.retab.com",
  });
  
  console.log(JSON.stringify(automation, null, 2));
  */
  
}

main().catch(console.error);