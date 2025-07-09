import { config } from 'dotenv';
import { Retab } from '@retab/node';



async function main() {
  config({ path: '../../../.env.production' });
  
  const reclient = new Retab();
  
  const automation = await reclient.processors.automations.outlook.create({
    name: "Outlook Automation",
    processor_id: "proc_o4dtLxizT0kDAjeKuyVLA",
    webhook_url: "https://your-server.com/webhook",
  });
  
  console.log(JSON.stringify(automation, null, 2));
  
}

main().catch(console.error);