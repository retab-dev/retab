import { config } from 'dotenv';
import { Retab } from '@retab/node';



async function main() {
  config({ path: '../../../.env.local' });
  
  const reclient = new Retab();
  
  const automation = await reclient.processors.automations.endpoints.get({
    endpoint_id: "endp_jKH0WW6X1dD3wtWbnuuXQ",
  });
  
  console.log(JSON.stringify(automation, null, 2));
  
}

main().catch(console.error);