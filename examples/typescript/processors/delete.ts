import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';



async function main() {
  config();
  
  const reclient = new AsyncRetab();
  
  await reclient.processors.delete("proc_lRLewYl5kVAmbEeW1AuuG");
  
}

main().catch(console.error);