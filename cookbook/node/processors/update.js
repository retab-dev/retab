import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const reclient = new AsyncRetab();

const processor = await reclient.processors.update({
  processor_id: "proc_F0FE8DFqyouQdZXDTWRg0",
  name: "Invoice Processor Updated",
});

console.log(JSON.stringify(processor, null, 2));