import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const reclient = new AsyncRetab();

const processor = await reclient.processors.get("proc_t51R0NeVxvWlBFEJJNVAd");

console.log(JSON.stringify(processor, null, 2));