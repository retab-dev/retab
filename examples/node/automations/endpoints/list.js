import { config } from 'dotenv';
import { Retab } from '@retab/node';

// Load environment variables
config({ path: '../../../.env.production' });

const reclient = new Retab();

const automations = await reclient.processors.automations.endpoints.list({
  processor_id: "proc_o4dtLxizT0kDAjeKuyVLA",
});

console.log(JSON.stringify(automations, null, 2));