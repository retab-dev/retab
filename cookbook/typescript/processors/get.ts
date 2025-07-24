import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';



async function main() {
  config();
  
  const reclient = new AsyncRetab();
  
  const processor = await reclient.processors.get("proc_Vq01vCzOnJctSkkrilzmJ");
  
  console.log(JSON.stringify(processor, null, 2));
  
}

main().catch(console.error);