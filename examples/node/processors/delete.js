import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const reclient = new AsyncRetab();

await reclient.processors.delete("proc_F0FE8DFqyouQdZXDTWRg0");