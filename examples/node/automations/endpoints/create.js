import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const reclient = new AsyncRetab();

console.log("Note: Endpoint automations are not yet fully implemented in the Node.js SDK");
console.log("This is a placeholder example showing the intended structure");

// This would be the intended usage once endpoint automations are fully implemented:
/*
console.log("making a query to create an endpoint");
const automation = await reclient.processors.automations.endpoints.create({
  name: "Endpoint Automation 2",
  processor_id: "proc_o4dtLxizT0kDAjeKuyVLA",
  webhook_url: "https://your-server.com/webhook",
});

console.log(JSON.stringify(automation, null, 2));
*/