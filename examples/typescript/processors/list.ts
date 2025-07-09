import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';



async function main() {
  config();
  
  const reclient = new AsyncRetab();
  
  const processors = await reclient.processors.list();
  
  console.log(JSON.stringify(processors, null, 2));
  
}

main().catch(console.error);