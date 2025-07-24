import { config } from 'dotenv';
import { Retab } from '@retab/node';

config({ path: '../../../.env.local' });

const reclient = new Retab();

const automation = await reclient.processors.automations.endpoints.update({
  endpoint_id: "endp_jKH0WW6X1dD3wtWbnuuXQ",
  name: "Endpoint Automation Updated",
});

console.log(JSON.stringify(automation, null, 2));