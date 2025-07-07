import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';

config();

const reclient = new AsyncRetab();

const processors = await reclient.processors.list();

console.log(JSON.stringify(processors, null, 2));