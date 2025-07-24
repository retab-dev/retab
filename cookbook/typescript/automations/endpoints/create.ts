import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';



async function main() {
  // Load environment variables  
  config({ path: '../../../.env.local' });
  
  const reclient = new AsyncRetab();
  
  console.log("making a query to create an endpoint");
  const automation = await reclient.processors.automations.endpoints.create({
    name: "Endpoint Automation 2",
    processor_id: "proc_o4dtLxizT0kDAjeKuyVLA",
    webhook_url: "https://your-server.com/webhook",
  });
  
  console.log(JSON.stringify(automation, null, 2));
  
}

main().catch(console.error);