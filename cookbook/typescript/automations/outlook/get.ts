import { config } from 'dotenv';
import { Retab } from '@retab/node';



async function main() {
  config({ path: '../../../.env.production' });
  
  const reclient = new Retab();
  
  const automation = await reclient.processors.automations.outlook.get({
    outlook_id: "outlook_DDw72RgWgIfKFXGaWXcGu",
  });
  
  console.log(JSON.stringify(automation, null, 2));
  
}

main().catch(console.error);